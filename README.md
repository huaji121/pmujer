# Pmujer

像素风格 2D 平台跳跃游戏，使用 Go + SDL3 开发。

## 操作

| 按键 | 功能 |
|------|------|
| A / D | 左移 / 右移 |
| W / J | 跳跃（支持二段跳） |
| R | 死亡后重生 |
| F3 | 切换 Debug 模式（显示碰撞箱） |
| ESC | 退出 |

## 构建与运行

```bash
go run ./src/          # 编译并运行
bash start.sh          # 同上
go build -o pmujer.exe ./src/  # 构建可执行文件
```

需要 Go 1.21+。SDL3/SDL3_image 通过 go-sdl3 内嵌，无需额外安装。

## 项目结构

```
src/
├── main.go              # 游戏入口、主循环、关卡布局
├── tilemap/             # Tile 系统（可扩展注册）
│   ├── tilemap.go       # 核心：碰撞检测、渲染
│   ├── tile_bricks.go   # 砖块定义
│   └── tile_spike.go   # 尖刺定义（凸多边形碰撞箱）
├── player/              # 玩家：移动、跳跃、物理、碰撞
├── camera/              # 帧率无关跟随摄像机
├── particle/            # 粒子效果系统
└── audio/               # WAV 音效播放
assets/
├── textures/            # 16×16 像素贴图
└── sounds/              # 音效文件
```
