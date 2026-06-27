# go-sdl3 API 参考

> 基于 `github.com/Zyko0/go-sdl3 v0.1.1`，SDL3 C API 的 Go 绑定（通过 purego 调用，无需 CGO）。

## 1. 初始化与生命周期

```go
import (
    "github.com/Zyko0/go-sdl3/bin/binsdl" // 嵌入式 SDL3 二进制
    "github.com/Zyko0/go-sdl3/bin/binimg" // 嵌入式 SDL3_image 二进制
    "github.com/Zyko0/go-sdl3/sdl"
    "github.com/Zyko0/go-sdl3/img"
)

func main() {
    defer binsdl.Load().Unload()
    defer binimg.Load().Unload()
    defer sdl.Quit()

    if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
        panic(err)
    }
    // ... 游戏代码 ...
}
```

- `binsdl.Load()` 将嵌入的 SDL3 DLL 解压到临时目录并调用 `sdl.LoadLibrary()`，无需系统安装 SDL3
- `binimg.Load()` 同理加载 SDL3_image（PNG/JPG 等支持）
- `sdl.Init()` 参数为 `InitFlags` 位掩码，常用 `sdl.INIT_VIDEO`（自动包含 `INIT_EVENTS`），需要音频时加 `sdl.INIT_AUDIO`
- 包还有 `binttf`（SDL3_ttf）、`binmix`（SDL3_mixer），用法相同
- defer 顺序：先 `binsdl`/`binimg`，再 `sdl.Quit()`；实际执行时 `Quit` 先于库卸载

## 2. 窗口与渲染器

```go
window, renderer, err := sdl.CreateWindowAndRenderer("标题", 960, 720, 0)
defer renderer.Destroy()
defer window.Destroy()
```

第 4 个参数为 `WindowFlags`，传 `0` 表示默认。

## 3. 主循环

```go
sdl.RunLoop(func() error {
    var event sdl.Event
    for sdl.PollEvent(&event) {
        switch event.Type {
        case sdl.EVENT_QUIT:
            return sdl.EndLoop
        case sdl.EVENT_KEY_DOWN:
            key := event.KeyboardEvent()  // union 类型，必须用方法访问
            if key.Scancode == sdl.SCANCODE_ESCAPE {
                return sdl.EndLoop
            }
        }
    }
    // update + render
    return nil
})
```

- `RunLoop` 内部就是 `for { if err := fn(); err != nil { ... } }`
- 返回 `sdl.EndLoop` 正常退出，返回其他 error 会传播出去

## 4. 事件系统

`sdl.Event` 是 union 类型，通过**类型化方法**访问具体事件：

```go
event.KeyboardEvent()    → *KeyboardEvent   // .Type, .Scancode, .Key, .Mod
event.MouseMotionEvent() → *MouseMotionEvent
event.MouseButtonEvent() → *MouseButtonEvent
event.WindowEvent()      → *WindowEvent
event.QuitEvent()        → *QuitEvent
// ... 共 20+ 种，见 sdl/glue.go
```

**键盘事件关键字段：**
```go
type KeyboardEvent struct {
    Type     EventType  // sdl.EVENT_KEY_DOWN / EVENT_KEY_UP
    Scancode Scancode   // 物理按键码
    Key      Keycode    // 虚拟按键码
    Mod      Keymod     // 修饰键状态
}
```

## 5. 键盘状态轮询

```go
keys := sdl.GetKeyboardState()  // 返回 []bool，按 Scancode 索引
if keys[sdl.SCANCODE_A] { /* A 键按下 */ }
if keys[sdl.SCANCODE_W] { /* W 键按下 */ }
if keys[sdl.SCANCODE_F3] { /* F3 键按下 */ }
```

比事件驱动更适合实时游戏输入。注意需要边沿检测时（如跳跃），需自行记录上一帧状态：

```go
jumpPressed := keys[sdl.SCANCODE_W]
if jumpPressed && !wasPressed {
    // 只在刚按下时触发
}
wasPressed = jumpPressed
```

## 6. 纹理

### 加载

```go
// 从文件加载（需 binimg，支持 PNG/JPG/BMP 等）
texture, err := img.LoadTexture(renderer, "path/to/image.png")
defer texture.Destroy()

// 从 Surface 创建
surface, _ := img.Load("path.png")
texture, _ := renderer.CreateTextureFromSurface(surface)
surface.Destroy()
```

### 属性与渲染

```go
texture.W  // int32，像素宽度
texture.H  // int32，像素高度

srcRect := sdl.FRect{X: 0, Y: 0, W: 16, H: 16}  // 源区域（nil = 整张）
dstRect := sdl.FRect{X: 100, Y: 100, W: 48, H: 48} // 目标区域
renderer.RenderTexture(texture, &srcRect, &dstRect)
```

### 透明度

```go
texture.SetAlphaMod(128)  // 0=全透明, 255=不透明
```

### 缩放模式

```go
texture.SetScaleMode(sdl.SCALEMODE_NEAREST)  // 像素风
texture.SetScaleMode(sdl.SCALEMODE_LINEAR)   // 平滑

// 设置渲染器默认（新建纹理自动继承）
renderer.SetDefaultTextureScaleMode(sdl.SCALEMODE_NEAREST)
```

## 7. 绘制

```go
renderer.SetDrawColor(r, g, b, a uint8)
renderer.Clear()                    // 用当前颜色清屏
renderer.Present()                  // 提交到屏幕（帧末尾调用）

renderer.RenderLine(x1, y1, x2, y2 float32)        // 直线
renderer.RenderRect(rect *FRect)                    // 矩形轮廓
renderer.RenderFillRect(rect *FRect)                // 填充矩形
```

注意：矩形轮廓方法是 `RenderRect`，不是 `DrawRect`。

## 8. 音频（SDL3 核心 API）

SDL3 音频使用 `AudioStream` 模型，无需 SDL3_mixer：

```go
// 1. 打开默认播放设备
spec := sdl.AudioSpec{Format: sdl.AUDIO_S16, Channels: 2, Freq: 44100}
devID, _ := sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDevice(&spec)
defer devID.Close()

// 2. 创建音频流并绑定到设备
stream, _ := sdl.CreateAudioStream(&spec, &spec)
defer stream.Destroy()
devID.BindAudioStream(stream)

// 3. 加载 WAV 并播放
data, _ := sdl.LoadWAV("jump.wav", &spec)  // 返回 []byte
stream.PutData(data)                         // 立即播放，可叠加
```

- `sdl.LoadWAV` 返回 `([]byte, error)`，数据在函数返回前已拷贝到 Go 内存
- `PutData` 是非阻塞的，多次调用会混合播放
- 需要 `sdl.INIT_AUDIO` 初始化标志

## 9. 时间与 Delta Time

```go
ticks := sdl.Ticks()             // uint64，毫秒（自 SDL 初始化）
ticksNS := sdl.TicksNS()         // uint64，纳秒

// 典型 delta time 计算
dt := float32(sdl.Ticks()-lastTicks) / 1000.0
lastTicks = sdl.Ticks()

// 帧率无关 lerp
lerpFactor := float32(1.0 - math.Exp(float64(-speed * dt)))
camX += (targetX - camX) * lerpFactor
```

## 10. 核心类型

```go
type FRect struct { X, Y, W, H float32 }   // 浮点矩形（渲染常用）
type Rect  struct { X, Y, W, H int32 }     // 整数矩形
type FPoint struct { X, Y float32 }         // 浮点点
type Point struct { X, Y int32 }            // 整数点
```

## 11. 与 SDL2 的主要差异

| 方面 | SDL2 (go-sdl2) | SDL3 (go-sdl3) |
|------|----------------|----------------|
| CGO | 需要 | 不需要（purego） |
| 事件访问 | `event.Key.Keysym.Scancode` | `event.KeyboardEvent().Scancode` |
| 主循环 | 手动 `for` | `sdl.RunLoop()` + `sdl.EndLoop` |
| 纹理渲染 | `renderer.Copy()` | `renderer.RenderTexture()` |
| 矩形轮廓 | `renderer.DrawRect()` | `renderer.RenderRect()` |
| 直线 | `renderer.DrawLine()` | `renderer.RenderLine()` |
| 错误返回 | `(int, error)` 或 `nil` | `error`（nil = 成功） |
| 布尔返回 | `int` (0/1) | `bool` |
| 音频 | `mixer.LoadWAV` + `PlayChannel` | `sdl.LoadWAV` + `stream.PutData` |
| 库加载 | CGO 链接 | `binsdl.Load()` 嵌入式 |

## 12. 包结构速查

```
github.com/Zyko0/go-sdl3/
├── sdl/          # SDL3 核心（窗口、渲染、事件、音频、...）
├── img/          # SDL3_image（LoadTexture, Load, SavePNG, ...）
├── ttf/          # SDL3_ttf（字体渲染）
├── mixer/        # SDL3_mixer（高级音频混合，SDL3 API 已不同于 SDL2）
├── bin/
│   ├── binsdl/   # 嵌入 SDL3.dll/.so/.dylib
│   ├── binimg/   # 嵌入 SDL3_image
│   ├── binttf/   # 嵌入 SDL3_ttf
│   └── binmix/   # 嵌入 SDL3_mixer
└── internal/     # 内部工具
```
