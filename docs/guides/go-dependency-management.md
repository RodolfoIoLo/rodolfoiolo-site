# Go 依赖管理规范

> 适用范围：本仓库 `server/` 模块（Go 1.25）。
> 依据：`go help mod` / `go help get` / `go help mod tidy` 等官方文档原文，以及本项目 [ADR-0006](../adr/0006-toolchain-baseline-2026.md) 的版本治理规则。
> 本文件是参考指南，不参与 `docs/README.md` 定义的文档权威顺序。

---

## 一、三条心智模型

**1. 真相只有一个：源码里的 `import`。**

`go.mod` 是「由 import 推导出的声明」，`go.sum` 是「这些声明的指纹」。三者的关系是单向的：

```
源码 import  ──推导──▶  go.mod  ──指纹──▶  go.sum
   （真相）            （声明）          （防篡改）
```

所以**永远不要手改 `go.mod` 里的依赖版本**，也不要手改 `go.sum`。要改就从源码或 `go get` 入手，让工具推导。

**2. Go 没有 lockfile —— 这是个容易带错的习惯。**

从 pnpm/npm 过来的人会找 `go.sum` 当 lockfile，它不是。

|          | pnpm             | Go                                |
| -------- | ---------------- | --------------------------------- |
| 声明依赖 | `package.json`   | `go.mod`                          |
| 锁定版本 | `pnpm-lock.yaml` | **`go.mod` 本身就是版本决策记录** |
| 锁定内容 | 哈希             | `go.sum`                          |

`go.mod` 里 `require foo v1.2.3` 的语义是**最低版本要求**，不是精确锁定。实际选中的版本由最小版本选择（MVS）决定——如果另一个依赖要求 `foo v1.3.0`，最终用的就是 v1.3.0。`go.sum` 锁的是「内容没被篡改」，不是「版本选择」。

**3. 默认 readonly。**

Go 1.16 起，`go build` / `go test` **不会**自动修改 `go.mod`。缺依赖就直接报错，让你自己决定。所以「整理」必须显式执行。

---

## 二、命令清单

### 2.1 引入与调整（会改 `go.mod`）

| 命令                              | 作用                                     | 什么时候用                                       |
| --------------------------------- | ---------------------------------------- | ------------------------------------------------ |
| `go get pkg@v1.2.3`               | 加/改 require、下载、写 `go.sum`         | **新增依赖首选**。本项目约定精确版本             |
| `go get pkg@latest`               | 取最新                                   | 只在明确要最新时                                 |
| `go get pkg@none`                 | 移除依赖，并降级依赖它的模块             | 彻底不要某模块时（比手删 require 干净）          |
| `go get -u pkg`                   | 升级 pkg **及其依赖**到最新 minor/patch  | 主动升级，注意会连带抖动                         |
| `go get -u=patch pkg`             | 只升 patch                               | 安全升级。注意写法是 `-u=patch`，不是 `-u patch` |
| `go get -tool pkg`                | 把命令行工具记入 `go.mod` 的 `tool` 指令 | 固定代码生成器/CLI 版本（Go 1.24+）              |
| `go get go@latest`                | 提升 `go` 指令                           | 升语言版本                                       |
| `go get toolchain@patch`          | 升 `toolchain` 指令到当前工具链最新补丁  | 安全补丁                                         |
| `go mod edit -go=1.25.0`          | 直接改 `go` 指令                         | 建仓固定                                         |
| `go mod edit -toolchain=go1.25.3` | 直接改 `toolchain` 指令                  | 建仓固定                                         |

`go` 与 `toolchain` 两条指令的分工（本项目 ADR-0006 已明确）：`go` 表达**代码使用的最低语言语义**，`toolchain` 表达**默认构建工具链**。分开是为了不让「语言特性版本」和「带安全修复的补丁工具链」互相绑架。

### 2.2 对账（以 import 为准重写声明）

| 命令                       | 作用                                                                 |
| -------------------------- | -------------------------------------------------------------------- |
| `go mod tidy`              | 扫全部源码 import：补缺失、删多余、重算 `// indirect`、补全 `go.sum` |
| `go mod tidy -diff`        | **只打印差异、不改文件**，有差异时非零退出 → CI 用                   |
| `go mod tidy -v`           | 打印删掉了哪些模块                                                   |
| `go mod tidy -go=1.25.0`   | 同时更新 `go` 指令                                                   |
| `go mod tidy -compat=1.24` | 保留指定旧版本所需的额外校验和                                       |

`-diff` 是本地预览「tidy 打算改什么」的最佳工具，比改完再 `git diff` 更安全。

### 2.3 只读检查（都不改文件）

| 命令                   | 回答什么问题                         |
| ---------------------- | ------------------------------------ |
| `go mod verify`        | 本地缓存里的依赖，内容有没有被改过？ |
| `go list -m all`       | 最终选中了哪些模块、哪些版本？       |
| `go mod why -m X`      | 我为什么需要模块 X？                 |
| `go mod why pkg`       | 我为什么需要包 pkg？                 |
| `go mod graph`         | 完整依赖图长什么样？                 |
| `go list -m -json all` | 机器可读的版本信息（脚本用）         |

**`go mod verify` 的准确语义**（官方原文）：检查**已下载到本地模块缓存**的依赖，其内容自下载以来有没有被修改过。全部未改动则打印 `all modules verified`，否则报告哪些模块被改并返回非零退出码。

注意它检查的是**本地缓存**，不是网络、不是 `go.sum` 文件本身。它的用途是：审计缓存完整性、CI 门禁、怀疑缓存被污染时定位。

### 2.4 搬运（把代码搬到本地）

| 命令                    | 作用                                                            |
| ----------------------- | --------------------------------------------------------------- |
| `go mod download`       | 预热本地缓存。无参数时下载主模块显式 require 的模块（Go 1.17+） |
| `go mod download -json` | 输出每个模块的路径、`Sum`、`GoModSum`，脚本用                   |
| `go mod vendor`         | 生成 `vendor/`，用于离线构建或依赖审计                          |

---

## 三、场景决策表

| 场景                         | 命令序列                                                                    |
| ---------------------------- | --------------------------------------------------------------------------- |
| **新加依赖**                 | 写 `import` → `go mod tidy`；或 `go get pkg@vX` → 写代码 → `go mod tidy`    |
| **删依赖**                   | 删 `import` → `go mod tidy`                                                 |
| **升级某依赖**               | `go get pkg@vX` → `go mod tidy` → `go test ./...`                           |
| **检查 go.mod 是否干净**     | `go mod tidy -diff`，有输出就是脏的                                         |
| **查某依赖从哪来**           | `go mod why -m X`                                                           |
| **本地能跑、CI 报缺 go.sum** | `go mod tidy`，然后提交 `go.mod` + `go.sum`                                 |
| **怀疑缓存被污染**           | `go mod verify`；失败则 `go clean -modcache` 后重新下载                     |
| **离线构建 / 依赖审计**      | `go mod vendor`，并把 `vendor/` 一并提交                                    |
| **固定工具版本**             | `go get -tool pkg@vX`（写入 `tool` 指令，Go 1.24+）                         |
| **漏洞扫描**                 | `go install golang.org/x/vuln/cmd/govulncheck@latest` → `govulncheck ./...` |
| **查最终选中版本**           | `go list -m all`                                                            |

---

## 四、本项目标准流程

### 4.1 新增依赖

```powershell
PS> Set-Location 'server'
PS> go get github.com/xxx/yyy@v1.2.3   # 精确版本，不写 @latest
# ... 在源码里 import 并使用它 ...
PS> go mod tidy
PS> go mod verify
PS> go test ./...
PS> Set-Location '..'
```

提交范围：`server/go.mod` + `server/go.sum`（ADR-0006 第 5 条：`go.sum` 必须提交）。

### 4.2 删除依赖

```powershell
PS> # 删掉源码里的 import
PS> Set-Location 'server'
PS> go mod tidy
PS> go mod verify
PS> Set-Location '..'
```

不要手删 `go.mod` 里的 `require` 行——`// indirect` 那些条目之间有关联，手删容易留下半截状态。

### 4.3 升级依赖

| 级别  | 做法                                         | 约束                                 |
| ----- | -------------------------------------------- | ------------------------------------ |
| patch | `go get pkg@vX.Y.Z` 或 `go get -u=patch pkg` | 过完整 CI 后合并（ADR-0006 第 2 条） |
| minor | 单独分支 + ADR                               | 禁止混入业务功能 PR（第 3 条）       |
| major | 单独分支 + ADR                               | 同上                                 |

Dependabot 每周提 PR，人工审后再合。紧急安全更新不等待季度窗口。

---

## 五、CI 门禁

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: server/go.mod
    cache-dependency-path: server/go.sum

- run: go mod download # 预热缓存，让后面的步骤快且可离线
- run: go mod verify # 缓存完整性
- run: go mod tidy -diff # go.mod 漂移检查：改完依赖忘了 tidy 会在这里挂
- run: go test -race ./...
```

`go mod tidy -diff` 是这套门禁里最有价值的一条：它把「`go.mod` 是否与源码同步」变成了一个可自动判定的布尔值。

---

## 六、环境变量

当前本机实际值（`go env`）：

| 变量                                    | 当前值                            | 说明                                               |
| --------------------------------------- | --------------------------------- | -------------------------------------------------- |
| `GOPROXY`                               | `https://proxy.golang.org,direct` | **建议改**，见下                                   |
| `GOSUMDB`                               | `sum.golang.org`                  | 校验和数据库                                       |
| `GOTOOLCHAIN`                           | `auto`                            | 按 `go.mod` 的 `toolchain` 指令自动切换/下载工具链 |
| `GOMODCACHE`                            | `C:\Users\Leonov\go\pkg\mod`      | 模块缓存位置                                       |
| `GOFLAGS`                               | 空                                | 不要轻易设 `-mod=mod`                              |
| `GOPRIVATE` / `GONOPROXY` / `GONOSUMDB` | 空                                | 私有仓库需要配置                                   |

### 建议调整

`proxy.golang.org` 在国内访问不稳定（本次实测出现过 `Bad Gateway`）。改成国内镜像：

```powershell
PS> go env -w GOPROXY=https://goproxy.cn,direct
```

`go env -w` 会写入用户级配置文件，对所有项目生效，不需要每次设环境变量。

### 几个关键变量的使用场景

- **`GOTOOLCHAIN=auto`**（默认）：Go 会读 `go.mod` 的 `toolchain go1.25.3`，如果本地版本不满足就**自动下载**对应工具链。这既保证了一致性，也是「为什么 CI 上会突然下载工具链」的原因。想强制只用本地版本：`GOTOOLCHAIN=local`。
- **`GOFLAGS=-mod=mod`**：让构建命令恢复「自动改 `go.mod`」的旧行为。**不建议**——它会掩盖「忘了 tidy」的问题，让 `go.mod` 悄悄偏离。
- **`GOPRIVATE=*.corp.example.com`**：私有仓库不走 proxy、不查 sumdb。它等价于同时设置 `GONOPROXY` 和 `GONOSUMDB`。
- **`-mod=vendor`**：当 `go.mod` 的 `go` ≥ 1.14 且存在 `vendor/modules.txt` 时**自动启用**。这意味着有 `vendor/` 的项目，改依赖后必须重跑 `go mod vendor`，否则构建用的是旧代码。

---

## 七、常见坑

1. **`go get` 之后代码还没 `import`，就顺手 `go mod tidy`** → 刚加的依赖被当作「未使用」删掉。先写代码，再 tidy。

2. **手写 `// indirect` 依赖的版本** → 下次 `tidy` 直接覆盖。ADR-0006 第 6 条已经禁止：由包管理器解析，以生成文件为准。

3. **`go.sum` 没提交** → 队友和 CI 报 `missing go.sum entry`。这是最常见的「我本地能跑啊」。

4. **`go get -u` 无脑全升** → 它会升级指定包**及其依赖**，依赖图整体抖动，可能引入不相关变更。升级单个包时优先 `go get pkg@vX.Y.Z` 指定版本。

5. **以为 `go mod tidy` 是「下载依赖」** → 不是。下载是 `go get` / `go mod download` / 构建时自动进行的事。`tidy` 只整理声明。

6. **手删 `go.mod` 里「看着没用」的依赖** → `tidy` 会「假装所有 build tag 都打开」，所以会保留只在 Windows 或特定 tag 下用到的依赖。手删会让其他平台构建失败。

7. **用 `go get` 安装命令行工具** → 那是 Go 1.16 之前的用法。现在用 `go install pkg@version`（它忽略当前目录的 `go.mod`），或者用 `go get -tool` 把它固定进项目。

8. **改了依赖忘了重跑 `go mod vendor`** → 有 `vendor/` 的项目会自动用 `-mod=vendor`，构建读的是 `vendor/` 里的旧代码。

---

## 八、与 ADR-0006 的对应关系

| ADR-0006 规则                                         | 落地命令                                                          |
| ----------------------------------------------------- | ----------------------------------------------------------------- |
| 2. Dependabot 每周提 PR，补丁版本过 CI 后合并         | `go get pkg@vX.Y.Z` → `go test ./...`                             |
| 3. major/minor 升级单独建分支和 ADR                   | 分支 + `go get` + ADR 文档                                        |
| 5. `go.sum` 必须提交；CI 用 frozen install 和漂移检查 | `git add server/go.sum`；CI `go mod verify` + `go mod tidy -diff` |
| 6. 不手写间接依赖版本，以生成文件为准                 | 禁止手改 `go.mod` 的 `// indirect` 段                             |
| Go 语言版本 / 工具链分离                              | `go mod edit -go=1.25.0` 与 `-toolchain=go1.25.3` 分开设置        |

---

## 九、一页速查

```
新增依赖      go get pkg@v1.2.3  →  写代码  →  go mod tidy
删除依赖      删 import  →  go mod tidy
升级依赖      go get pkg@vX.Y.Z  →  go mod tidy  →  go test ./...
改完自检      go mod tidy -diff          # 有输出 = 没整理干净
完整性审计    go mod verify              # 本地缓存有没有被改
查来源        go mod why -m X
查最终版本    go list -m all
预热缓存      go mod download
离线构建      go mod vendor
固定工具      go get -tool pkg@vX
```

**核心口诀：改 import → tidy → verify → 提交 go.mod 和 go.sum。**
