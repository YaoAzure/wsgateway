package limiter

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/YaoAzure/wsgateway/pkg/config"
	"github.com/samber/do/v2"
)

// TokenLimiterConfig 结构体用于配置 TokenLimiter。
// 这个配置结构体定义了令牌桶限流器的所有关键参数，支持动态容量增长的配置
type TokenLimiterConfig struct {
	// InitialCapacity 启动时的初始容量
	// 系统启动时令牌桶的初始大小，通常设置为较小的值以避免启动时的流量冲击
	// 例如：如果最大容量是1000，初始容量可以设置为100
	InitialCapacity int64 `yaml:"initialCapacity"`

	// MaxCapacity 最终的稳定容量  
	// 令牌桶能够达到的最大容量，也就是系统能够同时处理的最大并发数
	// 这个值应该根据系统的实际处理能力来设定
	MaxCapacity int64 `yaml:"maxCapacity"`

	// IncreaseStep 每次增加的容量步长
	// 在容量增长过程中，每次增加的令牌数量
	// 较小的步长可以让容量增长更平滑，较大的步长可以让系统更快达到最大容量
	IncreaseStep int64 `yaml:"increaseStep"`

	// IncreaseInterval 每次增加容量的时间间隔
	// 控制容量增长的速度，例如每30秒增加一次容量
	// 这个间隔应该根据系统的预热时间和负载特性来调整
	IncreaseInterval time.Duration `yaml:"increaseInterval"`
}

// TokenLimiter 通过令牌桶算法管理并发数，并支持容量的动态、逐步增长。
// 它是一个独立、可控生命周期的组件，并且是并发安全的。
//
// 设计理念：
// 1. 令牌桶算法：每个并发请求需要获取一个令牌，处理完成后归还令牌
// 2. 动态扩容：系统启动时使用较小容量，然后逐步增长到最大容量（类似于服务预热）
// 3. 并发安全：所有操作都是线程安全的，可以在多个goroutine中安全使用
// 4. 生命周期管理：支持优雅启动和关闭
type TokenLimiter struct {
	// config 限流器配置
	// 存储了所有的配置参数，在整个生命周期中保持不变
	config TokenLimiterConfig

	// currentCapacity 当前的实时容量
	// 使用atomic.Int64确保并发安全，记录当前令牌桶的实际容量
	// 这个值会从InitialCapacity逐步增长到MaxCapacity
	currentCapacity atomic.Int64

	// tokens 用作令牌桶
	// 这是一个带缓冲的channel，缓冲区大小等于MaxCapacity
	// channel中的每个元素代表一个可用的令牌
	// 获取令牌就是从channel中读取，归还令牌就是向channel中写入
	tokens chan struct{}

	// 组件内部的 context，用于通过 Close 方法从外部控制其生命周期。
	// ctx 内部上下文，当调用Close()方法时会被取消
	// 用于通知所有相关的goroutine停止运行
	ctx context.Context

	// cancel 取消函数，调用它会取消上面的ctx
	// 这是实现优雅关闭的关键机制
	cancel context.CancelFunc
}

// NewTokenLimiter 使用指定的配置创建一个新的 TokenLimiter 实例。
// 它会对配置进行严格校验，如果配置无效，将返回错误。
//
// 创建过程包括：
// 1. 严格的参数校验：确保所有配置参数都是合理的
// 2. 初始化内部组件：创建context、channel、logger等
// 3. 填充初始令牌：根据InitialCapacity预先放入令牌
// 4. 记录初始化日志：便于监控和调试
//
// 参数校验规则：
// - MaxCapacity 必须为正数（系统必须能处理至少1个并发）
// - InitialCapacity 不能为负数（不能有负的初始容量）
// - InitialCapacity 不能大于 MaxCapacity（初始容量不能超过最大容量）
// - IncreaseStep 必须为正数（每次增长的步长必须大于0）
// - IncreaseInterval 必须为正数（增长间隔必须大于0）
//
// 返回值：
// - 成功时返回配置好的TokenLimiter实例和nil错误
// - 失败时返回nil和具体的错误信息
func NewTokenLimiter(i do.Injector) (*TokenLimiter, error) {
	cfg, err := do.Invoke[config.ServerConfig](i)
	if err != nil {
		panic(fmt.Errorf("获取 TokenLimiterConfig 失败: %w", err))
	}

	tlc := TokenLimiterConfig{
		InitialCapacity: cfg.Websocket.TokenLimiter.InitialCapacity,
		MaxCapacity:     cfg.Websocket.TokenLimiter.MaxCapacity,
		IncreaseStep:    cfg.Websocket.TokenLimiter.IncreaseStep,
		IncreaseInterval: time.Duration(cfg.Websocket.TokenLimiter.IncreaseInterval),
	}

	// 1. 严格校验参数，提供更具体的错误信息
	// 检查最大容量：这是系统能处理的最大并发数，必须为正数
	if tlc.MaxCapacity <= 0 {
		return nil, errors.New("配置错误: MaxCapacity 必须为正数")
	}
	
	// 检查初始容量：不能为负数，负数没有实际意义
	if tlc.InitialCapacity < 0 {
		return nil, errors.New("配置错误: InitialCapacity 不能为负数")
	}
	
	// 检查初始容量与最大容量的关系：初始容量不能超过最大容量
	if tlc.InitialCapacity > tlc.MaxCapacity {
		return nil, fmt.Errorf("配置错误: InitialCapacity (%d) 不能大于 MaxCapacity (%d)", tlc.InitialCapacity, tlc.MaxCapacity)
	}
	
	// 检查增长步长：每次增长的令牌数必须为正数
	if tlc.IncreaseStep <= 0 {
		return nil, errors.New("配置错误: IncreaseStep 必须为正数")
	}
	
	// 检查增长间隔：时间间隔必须为正数
	if tlc.IncreaseInterval <= 0 {
		return nil, errors.New("配置错误: IncreaseInterval 必须为正数")
	}

	// 2. 创建实例
	// 创建可取消的上下文，用于控制组件的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	
	// 初始化TokenLimiter实例
	l := &TokenLimiter{
		config: tlc,
		// 创建令牌桶channel，缓冲区大小为最大容量
		// 这样可以确保在最大容量下不会阻塞
		tokens: make(chan struct{}, tlc.MaxCapacity),
		ctx:    ctx,
		cancel: cancel,
	}

	// 3. 填充初始令牌
	// 根据初始容量向令牌桶中放入对应数量的令牌
	// 这些令牌代表系统启动时就可以处理的并发数
	for i := int64(0); i < tlc.InitialCapacity; i++ {
		l.tokens <- struct{}{}
	}
	// 原子性地设置当前容量，确保并发安全
	l.currentCapacity.Store(tlc.InitialCapacity)

	return l, nil
}

// StartRampUp 启动一个后台 goroutine，该 goroutine 会逐步增加令牌桶的容量。
// 调用者需要负责在独立的 goroutine 中运行此方法。
//
// 这是实现动态扩容的核心方法，它的工作原理：
// 1. 创建定时器：按照配置的时间间隔定期触发容量增长
// 2. 监听取消信号：支持外部和内部两种取消机制
// 3. 逐步增加容量：每次按照配置的步长增加令牌数量
// 4. 自动停止：当达到最大容量时自动退出
//
// 双重取消机制说明：
// - 外部ctx：通常与特定的请求或任务绑定，当该任务结束时取消
// - 内部ctx：与TokenLimiter的生命周期绑定，当调用Close()时取消
// 
// 使用场景：
// - 系统启动时调用，让系统逐步达到满负荷运行状态
// - 在流量高峰期前调用，提前准备更多的处理能力
// - 在系统恢复时调用，避免瞬间冲击
//
// 注意事项：
// - 此方法会阻塞运行，必须在独立的goroutine中调用
// - 可以安全地多次调用，但通常只需要调用一次
// - 当容量达到最大值时会自动退出，不需要手动停止
//
// 此方法提供了两种停止机制：
// 1. 外部传入的 ctx：当这个 ctx 被取消时（例如，与单个请求或临时任务绑定），goroutine 会退出。
// 2. 内部的 ctx：当调用 TokenLimiter 的 Close 方法时，内部 ctx 会被取消，goroutine 也会退出。
func (t *TokenLimiter) StartRampUp(ctx context.Context) {
	// 创建定时器，按照配置的间隔定期触发容量增长
	// 使用defer确保定时器资源被正确释放
	ticker := time.NewTicker(t.config.IncreaseInterval)
	defer ticker.Stop()


	// 主循环：持续监听各种信号并处理容量增长
	for {
		select {
		case <-ctx.Done(): // 监听来自方法参数的取消信号
			// 外部上下文被取消，通常是因为相关的任务或请求结束了
			return
			
		case <-t.ctx.Done(): // 监听来自组件内部的取消信号 (由Close触发)
			// 内部上下文被取消，通常是因为整个TokenLimiter要关闭了
			return
			
		case <-ticker.C: // 定时器触发，执行容量增长逻辑
			// 原子性地读取当前容量，确保并发安全
			current := t.currentCapacity.Load()
			
			// 检查是否已经达到最大容量
			if current >= t.config.MaxCapacity {
				// 容量已达到最大值，记录日志并退出
				// 这个goroutine的使命已经完成，可以安全退出了
				return
			}

			// 计算本次增长后的新容量
			// 确保不会超过最大容量限制
			newCapacity := current + t.config.IncreaseStep
			if newCapacity > t.config.MaxCapacity {
				// 如果计算出的新容量超过了最大容量，则设置为最大容量
				newCapacity = t.config.MaxCapacity
			}

			// 向令牌桶中添加增量令牌
			// 计算需要添加的令牌数量
			addedTokens := newCapacity - current
			
			// 逐个添加令牌到channel中
			// 每个struct{}{}代表一个可用的令牌
			for i := int64(0); i < addedTokens; i++ {
				t.tokens <- struct{}{}
			}
			
			// 原子性地更新当前容量
			t.currentCapacity.Store(newCapacity)
		}
	}
}

// Acquire 尝试获取一个令牌。
// 这是一个非阻塞操作。如果成功获取到令牌，返回 true；
// 如果当前没有可用令牌，立即返回 false。
//
// 工作原理：
// - 使用select语句的default分支实现非阻塞行为
// - 尝试从tokens channel中读取一个令牌（struct{}{}）
// - 如果channel中有令牌，立即返回true
// - 如果channel为空，立即返回false，不会阻塞等待
//
// 使用场景：
// - 在处理新的并发请求前调用，检查是否有可用的处理能力
// - 通常与Release()成对使用，形成完整的资源管理周期
//
// 返回值：
// - true：成功获取到令牌，可以继续处理请求
// - false：当前没有可用令牌，应该拒绝或延迟处理请求
//
// 注意事项：
// - 此方法是并发安全的，可以在多个goroutine中同时调用
// - 获取令牌后必须在适当的时候调用Release()归还令牌
// - 不要忘记处理获取失败的情况（返回false时）
func (t *TokenLimiter) Acquire() bool {
	select {
	case <-t.tokens:
		return true // 成功获取令牌
	default:
		return false // 令牌桶已空
	}
}

// Release 归还一个令牌。非阻塞。
//
// 工作原理：
// - 使用select语句的default分支实现非阻塞行为
// - 尝试向tokens channel中写入一个令牌（struct{}{}）
// - 如果channel未满，立即写入并返回true
// - 如果channel已满，立即返回false并记录警告日志
//
// 使用场景：
// - 在完成请求处理后调用，释放占用的处理能力
// - 与Acquire()成对使用，确保资源的正确管理
//
// 返回值：
// - true：成功归还令牌
// - false：归还失败，通常表示存在逻辑错误
//
// 异常情况处理：
// - 如果归还失败（返回false），通常意味着Release()的调用次数超过了Acquire()
// - 这种情况会记录警告日志，便于发现和修复代码中的逻辑错误
//
// 注意事项：
// - 此方法是并发安全的
// - 只有在成功调用Acquire()后才应该调用此方法
// - 不要重复归还同一个令牌
func (t *TokenLimiter) Release() bool {
	select {
	case t.tokens <- struct{}{}:
		return true
	default:
		// 这种情况理论上不应该发生，除非 Release 的调用次数超过了 Acquire。
		// 这通常意味着代码中存在逻辑错误。
		return false
	}
}

// Close 会取消组件内部的 context，从而通知所有由该 limiter 启动的后台 goroutine 停止。
// 这是一个优雅关闭的必要部分，应该在服务关闭时被调用。
// 这个方法是幂等的，可以安全地多次调用。
//
// 关闭过程：
// 1. 调用cancel()函数取消内部的context
// 2. 所有监听t.ctx.Done()的goroutine会收到取消信号
// 3. 相关的后台任务（如StartRampUp）会优雅地停止
// 4. 记录关闭日志
//
// 影响范围：
// - 正在运行的StartRampUp goroutine会停止容量增长
// - 不会影响已经获取的令牌，它们仍然有效
// - 不会影响Acquire()和Release()的正常使用
//
// 使用场景：
// - 应用程序关闭时
// - 服务重启时
// - 不再需要限流功能时
//
// 注意事项：
// - 此方法是幂等的，多次调用不会产生副作用
// - 调用后TokenLimiter仍然可以用于Acquire/Release操作
// - 但是StartRampUp功能会停止工作
func (t *TokenLimiter) Close() error {
	// 取消内部context，通知所有相关的goroutine停止
	t.cancel()
	
	return nil
}

// CurrentCapacity 返回限流器当前的实时容量。
// 这个方法是新增的，用于支持包外测试，让测试代码可以检查内部状态。
//
// 工作原理：
// - 使用atomic.Int64.Load()原子性地读取当前容量
// - 确保在并发环境下读取到的值是一致的
//
// 返回值：
// - 当前令牌桶的实际容量（不是可用令牌数量）
// - 这个值会从InitialCapacity逐步增长到MaxCapacity
//
// 使用场景：
// - 监控系统当前的处理能力
// - 测试代码验证容量增长是否正常
// - 调试和故障排查
//
// 注意事项：
// - 返回的是容量大小，不是当前可用的令牌数量
// - 可用令牌数量 = 容量大小 - 当前正在使用的令牌数量
// - 此方法是并发安全的
func (t *TokenLimiter) CurrentCapacity() int64 {
	return t.currentCapacity.Load()
}
