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
    // 加载顺序：先 SDL，再 image；defer 保证反序释放
    defer binsdl.Load().Unload()
    defer binimg.Load().Unload()
    defer sdl.Quit()

    if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
        panic(err)
    }
    // ... 游戏代码 ...
}
```

- `binsdl.Load()` 会将嵌入的 SDL3 DLL 解压到临时目录并调用 `sdl.LoadLibrary()`，无需系统安装 SDL3
- `binimg.Load()` 同理加载 SDL3_image（PNG/JPG 等支持）
- `sdl.Init()` 参数为 `InitFlags` 位掩码，常用 `sdl.INIT_VIDEO`（自动包含 `INIT_EVENTS`）
- 包还有 `binttf`（SDL3_ttf）、`binmix`（SDL3_mixer），用法相同

## 2. 窗口与渲染器

```go
// 一步创建窗口+渲染器
window, renderer, err := sdl.CreateWindowAndRenderer("标题", 960, 720, 0)
defer renderer.Destroy()
defer window.Destroy()

// 或分开创建
window, err := sdl.CreateWindow("标题", 960, 720, 0)
renderer, err := sdl.CreateRenderer(window, nil)
```

第 4 个参数为 `WindowFlags`，传 `0` 表示默认。

## 3. 主循环

go-sdl3 提供 `sdl.RunLoop()` 替代手动 `for` 循环，返回 `sdl.EndLoop` 即可退出：

```go
sdl.RunLoop(func() error {
    var event sdl.Event
    for sdl.PollEvent(&event) {
        switch event.Type {
        case sdl.EVENT_QUIT:
            return sdl.EndLoop
        case sdl.EVENT_KEY_DOWN:
            key := event.KeyboardEvent()  // 注意：用方法访问，不是字段
            if key.Scancode == sdl.SCANCODE_ESCAPE {
                return sdl.EndLoop
            }
        }
    }
    // update + render
    return nil
})
```

`RunLoop` 内部就是 `for { if err := fn(); err != nil { ... } }`，返回非 `EndLoop` 的 error 会传播出去。

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
    Scancode Scancode   // 物理按键码（sdl.SCANCODE_A = 4, _D = 7, _W = 26, _J = 13, ...）
    Key      Keycode    // 虚拟按键码
    Mod      Keymod     // 修饰键状态
    // ...
}
```

## 5. 键盘状态轮询

每帧获取整个键盘快照，按 `Scancode` 索引（`bool` 切片）：

```go
keys := sdl.GetKeyboardState()
if keys[sdl.SCANCODE_A] { /* A 键按下 */ }
if keys[sdl.SCANCODE_W] { /* W 键按下 */ }
```

比事件驱动更适合实时游戏输入。

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

### 属性

```go
texture.W  // int32，像素宽度
texture.H  // int32，像素高度
```

### 缩放模式

```go
// 设置单个纹理
texture.SetScaleMode(sdl.SCALEMODE_NEAREST)  // 像素风
texture.SetScaleMode(sdl.SCALEMODE_LINEAR)   // 平滑
texture.SetScaleMode(sdl.SCALEMODE_PIXELART) // SDL 3.4+ 像素风增强

// 设置渲染器默认（新建纹理自动继承）
renderer.SetDefaultTextureScaleMode(sdl.SCALEMODE_NEAREST)
```

### 渲染

```go
srcRect := sdl.FRect{X: 0, Y: 0, W: 16, H: 16}  // 源区域（nil = 整张）
dstRect := sdl.FRect{X: 100, Y: 100, W: 48, H: 48} // 目标区域
renderer.RenderTexture(texture, &srcRect, &dstRect)
```

## 7. 绘制

```go
renderer.SetDrawColor(r, g, b, a uint8)  // 设置清屏/画笔颜色
renderer.Clear()                          // 用当前颜色清屏
renderer.Present()                        // 提交到屏幕（帧末尾调用）

// 绘制几何图形
renderer.DrawLine(x1, y1, x2, y2 float32)
renderer.DrawRect(rect *FRect)
renderer.FillRect(rect *FRect)
```

## 8. 核心类型

```go
type FRect struct { X, Y, W, H float32 }   // 浮点矩形（渲染常用）
type Rect  struct { X, Y, W, H int32 }     // 整数矩形
type FPoint struct { X, Y float32 }         // 浮点点
type Point struct { X, Y int32 }            // 整数点
```

## 9. 时间

```go
ticks := sdl.Ticks()             // uint64，毫秒（自 SDL 初始化）
ticksNS := sdl.TicksNS()         // uint64，纳秒
counter := sdl.GetPerformanceCounter()
freq := sdl.GetPerformanceFrequency()

// 典型 delta time 计算
dt := float32(sdl.Ticks()-lastTicks) / 1000.0
lastTicks = sdl.Ticks()
```

## 10. 与 SDL2 的主要差异

| 方面 | SDL2 (go-sdl2) | SDL3 (go-sdl3) |
|------|----------------|----------------|
| CGO | 需要 | 不需要（purego） |
| 事件访问 | `event.Key.Keysym.Scancode` | `event.KeyboardEvent().Scancode` |
| 主循环 | 手动 `for` | `sdl.RunLoop()` + `sdl.EndLoop` |
| 纹理渲染 | `renderer.Copy()` / `CopyEx()` | `renderer.RenderTexture()` / `RenderTextureRotated()` |
| 错误返回 | `(int, error)` 或 `nil` | `error`（nil = 成功） |
| 布尔返回 | `int` (0/1) | `bool` |
| 库加载 | CGO 链接 | `binsdl.Load()` 嵌入式 / `sdl.LoadLibrary()` 外部 |

## 11. 包结构速查

```
github.com/Zyko0/go-sdl3/
├── sdl/          # SDL3 核心（Init, Window, Renderer, Event, ...）
├── img/          # SDL3_image（LoadTexture, Load, SavePNG, ...）
├── ttf/          # SDL3_ttf（字体渲染）
├── mixer/        # SDL3_mixer（音频混合）
├── bin/
│   ├── binsdl/   # 嵌入 SDL3.dll/.so/.dylib
│   ├── binimg/   # 嵌入 SDL3_image
│   ├── binttf/   # 嵌入 SDL3_ttf
│   └── binmix/   # 嵌入 SDL3_mixer
└── internal/     # 内部工具
```
