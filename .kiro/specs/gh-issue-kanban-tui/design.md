# Design Document

## Overview

**Purpose**: GitHub Issueをカンバンボード形式で管理するGitHub CLI拡張機能を提供する。`gh issue-tui`コマンドとしてターミナル上でIssueの閲覧・編集・作成を完結させる。

**Users**: GitHub上でIssue管理を行う開発者。ターミナルベースのワークフローを好み、ブラウザへの切り替えを最小化したいユーザー。

**Impact**: GitHub CLIのエコシステムにカンバンボード機能を追加し、Issue管理のターミナル体験を大幅に向上させる。

### Goals
- GitHub Projects V2のカスタムステータスに連動したカンバンボードでIssueを視覚的に一覧表示する
- Issue詳細の閲覧・プロパティ編集・新規作成をTUI上で完結させる
- Vimライクなキーバインドおよび矢印キーによるキーボードナビゲーションで効率的な操作を実現する
- ステータス間のIssue移動（左右コマンド）をサポートする
- GitHub CLI拡張機能としてゼロコンフィグでインストール・利用可能にする
- Project紐付けを設定ファイルに永続化し、次回以降の起動を自動化する
- カラム表示・非表示の切り替えで関心のあるステータスに集中できるようにする
- Issue詳細画面からプロパティをインラインで編集可能にする

### Non-Goals
- Pull Requestの管理機能
- 複数リポジトリの同時表示
- オフラインモード・ローカルキャッシュの永続化
- Issue テンプレートのサポート
- Projects V2のカスタムフィールド（Status以外）の編集

## Architecture

### Architecture Pattern & Boundary Map

Elmアーキテクチャ（Model→Update→View）をBubble Teaが提供する基盤上に構築し、UIレイヤーとインフラレイヤーをクリーンに分離する。

```mermaid
graph TB
    subgraph CLI
        Main[main - エントリポイント]
    end

    subgraph CLI Layer
        Prompt[CLI Prompt - Project Selection via huh]
    end

    subgraph UI Layer
        App[App Model]
        Board[Board Model]
        Detail[Detail Model]
        Editor[Editor Model]
        Filter[Filter Model]
        Help[Help Model]
    end

    subgraph Domain Layer
        IssueService[Issue Service]
        ProjectService[Project Service]
        RepoService[Repo Service]
        ConfigService[Config Service]
    end

    subgraph Infrastructure Layer
        GHClient[GitHub API Client]
        GHAuth[gh auth 統合]
        ConfigFile[設定ファイル .gh-tuissue.json]
    end

    Main --> ConfigService
    Main --> Prompt
    Main --> App
    Prompt --> ProjectService
    Prompt --> ConfigService
    App --> Board
    App --> Detail
    App --> Editor
    App --> Filter
    App --> Help
    Board --> IssueService
    Board --> ProjectService
    Detail --> IssueService
    Detail --> RepoService
    Editor --> IssueService
    Filter --> IssueService
    IssueService --> GHClient
    ProjectService --> GHClient
    RepoService --> GHClient
    GHClient --> GHAuth
    ConfigService --> ConfigFile
```

**Architecture Integration**:
- **Selected pattern**: Elmアーキテクチャ + レイヤード構成。Bubble Tea v2の標準パターンに準拠しつつ、GitHub API層を分離
- **Domain boundaries**: UI Model群はBubble Teaのtea.Modelインターフェースを実装。Domain ServiceはUI非依存でAPIアクセスを抽象化
- **New components rationale**: 各画面（Board/Detail/Editor/Filter/Help）を独立Modelとして分離し、並行開発を可能にする。Project選択はTUI起動前にCLI層のインラインプロンプト（`charmbracelet/huh`）で実行し、TUI内の画面遷移から分離する。ProjectServiceはGitHub Projects V2のステータス管理を担当。ConfigServiceは設定ファイルの読み書きを担当

### Technology Stack

| Layer | Choice / Version | Role in Feature | Notes |
|-------|------------------|-----------------|-------|
| Language | Go 1.22+ | 実装言語 | gh extensionの標準言語 |
| TUI Framework | Bubble Tea v2 | UI描画・イベント処理 | charm.land/bubbletea/v2 |
| Styling | Lip Gloss | ターミナルスタイリング | カラー・ボーダー・レイアウト |
| Components | Bubbles | 再利用コンポーネント | viewport, list, textinput等 |
| GitHub CLI統合 | go-gh v2 | 認証・APIクライアント | github.com/cli/go-gh/v2 |
| GitHub API | GraphQL + REST | Issue CRUD操作 | 読み取りGraphQL、更新REST補完 |
| Markdown | glamour | Markdown→ターミナル描画 | Issue本文・コメント表示 |
| CLI Prompt | charmbracelet/huh | TUI起動前のインラインプロンプト | Project紐付け確認・Project選択 |
| CLI Flags | cobra or pflag | コマンドライン引数処理 | --repo, --project, --config, --help, --version |
| Build/Release | gh-extension-precompile | マルチプラットフォームビルド | GitHub Actions |
| Config Storage | encoding/json | 設定ファイル永続化 | Go標準ライブラリ、外部依存なし |

> 詳細な技術選定の根拠は `research.md` を参照。

## System Flows

### Issue取得・カンバン表示フロー

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant App as App Model
    participant Board as Board Model
    participant PSvc as Project Service
    participant ISvc as Issue Service
    participant API as GitHub GraphQL

    U->>App: gh issue-tui 起動
    App->>PSvc: GetProject(owner, repo, projectNumber)
    PSvc->>API: GraphQL query - ProjectV2 fields and options
    API-->>PSvc: Project + Status field + options
    App->>PSvc: GetProjectItems(projectId)
    PSvc->>API: GraphQL query - ProjectV2 items with fieldValues
    API-->>PSvc: ProjectV2Items with status values
    PSvc-->>Board: []ProjectItem with status mapping
    Board-->>U: カンバンボード表示 - カスタムステータスカラム
```

### 初回起動・Project紐付けフロー

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant Main as main
    participant Cfg as ConfigService
    participant Prompt as CLI Prompt via huh
    participant PSvc as ProjectService
    participant API as GitHub GraphQL
    participant App as App Model

    Main->>Cfg: LoadConfig(repoRoot)
    alt 設定ファイル存在
        Cfg-->>Main: Config with projectNumber
        Main->>PSvc: GetProjectFields(projectNumber)
        alt Project有効
            PSvc-->>Main: ProjectInfo
            Main->>App: TUI起動（Project付き）
        else Project無効・削除済み
            PSvc-->>Main: Error
            Main->>Main: Project選択フローへフォールバック
        end
    else 設定ファイル不在
        Cfg-->>Main: nil
    end
    Main->>Prompt: Bind a GitHub Project? Yes/No
    U->>Prompt: Yes選択
    Prompt->>PSvc: ListProjects(owner, repo)
    PSvc->>API: GraphQL query - repositoryのprojectsV2
    API-->>PSvc: []ProjectSummary
    PSvc-->>Prompt: Project一覧
    Prompt-->>U: Select a project - 一覧表示
    U->>Prompt: Project選択
    Prompt->>Cfg: SaveConfig(repoRoot, projectNumber)
    Main->>App: TUI起動（選択されたProject付き）
    App-->>U: カンバンボード表示
```

Project選択はTUI起動前にCLIのインラインプロンプト（`charmbracelet/huh`のConfirm/Select）で実行される。`gh repo create`等のGitHub CLIと同様のUXを提供する。

### Issueステータス移動フロー

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant Board as Board Model
    participant PSvc as Project Service
    participant API as GitHub GraphQL

    U->>Board: ステータス移動コマンド - 左or右
    Board->>Board: 隣接ステータスのoptionIdを解決
    Board->>PSvc: MoveItemStatus(projectId, itemId, fieldId, optionId)
    PSvc->>API: updateProjectV2ItemFieldValue mutation
    alt 成功
        API-->>PSvc: Updated ProjectV2Item
        PSvc-->>Board: 更新結果
        Board-->>U: Issueを移動先カラムに再配置して表示
    else 失敗
        API-->>PSvc: Error
        PSvc-->>Board: エラー
        Board-->>U: エラーメッセージ表示 + 元のカラムに維持
    end
```

### Issue編集フロー

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant Detail as Detail Model
    participant Svc as Issue Service
    participant API as GitHub API

    U->>Detail: プロパティ編集操作
    Detail->>Svc: UpdateIssue(number, changes)
    Svc->>API: REST PATCH or GraphQL mutation
    alt 成功
        API-->>Svc: Updated Issue
        Svc-->>Detail: 更新結果
        Detail-->>U: 画面更新
    else 失敗
        API-->>Svc: Error
        Svc-->>Detail: エラー
        Detail-->>U: エラーメッセージ表示 + 変更前状態に復帰
    end
```

### Issue詳細インライン編集フロー

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant Detail as Detail Model
    participant RSvc as Repo Service
    participant ISvc as Issue Service
    participant API as GitHub API

    U->>Detail: プロパティ編集キー（l/a/m）
    Detail->>RSvc: メタデータ取得（ラベル/コラボレーター/マイルストーン）
    RSvc->>API: REST GET
    API-->>RSvc: メタデータ一覧
    RSvc-->>Detail: 選択肢リスト
    Detail-->>U: 選択UI表示（現在の値がプリセレクト）
    U->>Detail: 選択確定
    Detail->>ISvc: UpdateIssue(number, changes)
    ISvc->>API: REST PATCH
    alt 成功
        API-->>ISvc: Updated Issue
        ISvc-->>Detail: 更新結果
        Detail-->>U: Issue詳細画面更新
    else 失敗
        API-->>ISvc: Error
        ISvc-->>Detail: エラー
        Detail-->>U: エラーメッセージ + 元の状態維持
    end
```

## Requirements Traceability

| Requirement | Summary | Components | Interfaces | Flows |
|-------------|---------|------------|------------|-------|
| 1.1 | gh extension installでインストール可能 | Main, Build設定 | CLI | - |
| 1.2 | gh issue-tuiで起動 | Main | CLI | - |
| 1.3 | gh auth token認証 | GHClient, GHAuth | GitHubAPIClient | - |
| 1.4 | 未認証時エラー表示 | GHClient | GitHubAPIClient | - |
| 1.5 | リポジトリ自動検出 | RepoService | RepoResolver | - |
| 1.6 | --repoフラグ指定 | Main | CLI | - |
| 2.1 | カンバンボード表示 | Board | BoardModel | Issue取得フロー |
| 2.2 | Project定義カスタムステータスカラム | Board, ProjectService | BoardModel, ProjectService | Issue取得フロー |
| 2.3 | Issueカード情報表示 | Board | IssueCard | - |
| 2.4 | カラム内スクロール | Board | BoardModel | - |
| 2.5 | レスポンシブレイアウト | Board | BoardModel | - |
| 3.1 | フィルタリング | Filter, Board | FilterModel | - |
| 3.2 | ソート | Filter, Board | FilterModel | - |
| 3.3 | フィルタ条件表示 | Filter | FilterModel | - |
| 3.4 | フィルタ結果0件表示 | Board | BoardModel | - |
| 4.1 | Issue詳細ビュー表示 | Detail | DetailModel | - |
| 4.2 | 詳細情報（Markdown含む） | Detail | DetailModel | - |
| 4.3 | コメント一覧表示 | Detail | DetailModel | - |
| 4.4 | 詳細ビュースクロール | Detail | DetailModel | - |
| 5.1 | ステータス移動（左右コマンド） | Board, ProjectService | ProjectService | ステータス移動フロー |
| 5.2 | タイトル編集 | Editor | EditorModel | Issue編集フロー |
| 5.3 | 本文編集（$EDITOR） | Editor | EditorModel | Issue編集フロー |
| 5.4 | ラベル編集 | Detail, IssueService | IssueService | Issue編集フロー |
| 5.5 | アサイニー編集 | Detail, IssueService | IssueService | Issue編集フロー |
| 5.6 | マイルストーン編集 | Detail, IssueService | IssueService | Issue編集フロー |
| 5.7 | API失敗時ロールバック | IssueService | IssueService | Issue編集フロー |
| 5.8 | 編集後カンバン画面復帰 | App, Detail, Editor | AppModel | - |
| 6.1 | Issue新規作成フォーム | Editor | EditorModel | - |
| 6.2 | 本文入力（$EDITOR） | Editor | EditorModel | - |
| 6.3 | 作成後カンバン更新 | Board, IssueService | IssueService | - |
| 6.4 | 作成時メタデータ設定 | Editor | EditorModel | - |
| 7.1 | コメント追加（$EDITOR） | Detail, Editor | EditorModel | - |
| 7.2 | コメント一覧更新 | Detail | DetailModel | - |
| 7.3 | コメント投稿失敗時保持 | Editor | EditorModel | - |
| 8.1 | Vim + 矢印キーナビゲーション | App | KeyMap | - |
| 8.2 | Enter/Esc操作 | App | KeyMap | - |
| 8.3 | ヘルプ表示 | Help | HelpModel | - |
| 8.4 | q終了 + 画面クリア | App | KeyMap, tea.WithAltScreen | - |
| 8.5 | キーバインドヒント | App | StatusBar | - |
| 9.1 | リフレッシュ | Board, IssueService | IssueService | Issue取得フロー |
| 9.2 | ローディング表示 | Board | BoardModel | - |
| 9.3 | タイムアウト時キャッシュ表示 | IssueService | IssueService | - |
| 10.1 | 初回起動時Project紐付け確認CLIプロンプト | Main, CLIPrompt | huh.Confirm | 初回起動フロー |
| 10.2 | Yes選択時Project一覧選択CLIプロンプト | Main, CLIPrompt, ProjectService | huh.Select, ProjectService | 初回起動フロー |
| 10.3 | No選択時Open/Closed 2カラム表示 | Main, Board | AppModel, BoardModel | 初回起動フロー |
| 10.4 | 設定ファイル保存 | ConfigService | ConfigService | 初回起動フロー |
| 10.5 | 次回起動時設定自動読み込み | Main, ConfigService | ConfigService | 初回起動フロー |
| 10.6 | --projectフラグ優先 | Main | CLI | - |
| 10.7 | --configフラグで設定変更 | Main, CLIPrompt | CLI, huh.Confirm, huh.Select | 初回起動フロー |
| 10.8 | 無効Project時の再選択 | Main, CLIPrompt | ConfigService, ProjectService, huh | 初回起動フロー |
| 11.1 | カラム非表示操作 | Board | BoardModel | - |
| 11.2 | カラム表示復元操作 | Board | BoardModel | - |
| 11.3 | 非表示カラム数ステータスバー表示 | Board, App | BoardModel, StatusBar | - |
| 11.4 | 最低1カラム表示維持 | Board | BoardModel | - |
| 12.1 | 詳細画面ラベルインライン編集 | Detail, RepoService, IssueService | DetailModel, SelectModel | インライン編集フロー |
| 12.2 | 詳細画面アサイニーインライン編集 | Detail, RepoService, IssueService | DetailModel, SelectModel | インライン編集フロー |
| 12.3 | 詳細画面マイルストーンインライン編集 | Detail, RepoService, IssueService | DetailModel, SelectModel | インライン編集フロー |
| 12.4 | 編集完了後詳細画面復帰・更新反映 | Detail | DetailModel | インライン編集フロー |

## Components and Interfaces

| Component | Domain/Layer | Intent | Req Coverage | Key Dependencies | Contracts |
|-----------|-------------|--------|--------------|------------------|-----------|
| CLIPrompt | CLI | TUI起動前のProject紐付け確認・選択 | 10.1-10.3, 10.7-10.8 | ProjectService (P0), ConfigService (P0), huh (P0) | Service |
| App | UI | トップレベル画面遷移・キーバインド管理 | 5.8, 8.1-8.5, 11.3 | Board, Detail, Editor, Filter, Help (P0) | State |
| Board | UI | カンバンボード表示・ナビゲーション | 2.1-2.5, 3.4, 5.1, 9.2, 11.1-11.4 | IssueService (P0), ProjectService (P0), Filter (P1) | State |
| Detail | UI | Issue詳細表示・プロパティ操作・インライン編集 | 4.1-4.4, 5.4-5.6, 7.1-7.2, 12.1-12.4 | IssueService (P0), RepoService (P1) | State |
| Editor | UI | テキスト入力・外部エディタ起動 | 5.2-5.3, 6.1-6.4, 7.1, 7.3 | IssueService (P0) | State |
| Filter | UI | フィルタ・ソートUI | 3.1-3.3 | RepoService (P1) | State |
| Help | UI | キーバインドヘルプ表示 | 8.3 | なし | State |
| IssueService | Domain | Issue CRUD操作の抽象化 | 5.2-5.7, 6.3, 7.2, 9.1, 9.3, 12.1-12.3 | GHClient (P0) | Service |
| ProjectService | Domain | GitHub Projects V2ステータス管理 | 2.2, 5.1, 10.2 | GHClient (P0) | Service |
| RepoService | Domain | リポジトリ情報・メタデータ取得 | 1.5, 3.1, 12.1-12.3 | GHClient (P0) | Service |
| ConfigService | Domain | 設定ファイルの読み書き | 10.4-10.8 | ConfigFile (P0) | Service |
| GHClient | Infrastructure | GitHub API通信 | 1.3-1.4 | go-gh v2 (P0) | Service |

### UI Layer

#### App Model

| Field | Detail |
|-------|--------|
| Intent | トップレベルの画面遷移管理とグローバルキーバインドのディスパッチ |
| Requirements | 5.8, 8.1, 8.2, 8.3, 8.4, 8.5, 11.3 |

**Responsibilities & Constraints**
- 現在のアクティブ画面（Board/Detail/Editor/Filter/Help）の管理
- グローバルキーバインド（q終了、?ヘルプ、Escで戻る）の処理。Vimキーバインド（h/j/k/l）と矢印キー（←/↓/↑/→）の両方をサポート
- 画面下部のステータスバー・キーバインドヒントの描画
- 編集操作完了後のカンバンボード画面への確実な復帰（5.8）
- `tea.WithAltScreen()`オプションによるaltscreenモードの使用（終了時の画面クリア、8.4）
- 非表示カラム数のステータスバー表示（11.3）
- Project選択はTUI起動前にCLI層で完了済み。AppModelはprojectNumberを受け取るのみ

**Dependencies**
- Outbound: Board, Detail, Editor, Filter, Help — 各画面Modelへの遷移 (P0)

**Contracts**: State [x]

##### State Management

```go
type ViewState int

const (
    ViewBoard ViewState = iota
    ViewDetail
    ViewEditor
    ViewFilter
    ViewHelp
)

type AppModel struct {
    currentView    ViewState
    board          BoardModel
    detail         DetailModel
    editor         EditorModel
    filter         FilterModel
    help           HelpModel
    statusMsg      string
    width          int
    height         int
}
```

- Persistence: メモリ内のみ（永続化なし）
- Concurrency: Bubble Teaのシングルスレッドイベントループで管理

#### Board Model

| Field | Detail |
|-------|--------|
| Intent | Issue一覧をGitHub Projects V2のカスタムステータスカラムでカンバンボード形式で描画する |
| Requirements | 2.1, 2.2, 2.3, 2.4, 2.5, 3.4, 5.1, 9.2, 11.1, 11.2, 11.3, 11.4 |

**Responsibilities & Constraints**
- GitHub Projects V2のStatusフィールドのoptionsに基づくN個の動的カラムレイアウトでIssueカードを描画
- カラム間・カード間のカーソル移動（h/j/k/l および ←/↓/↑/→）
- ステータス移動コマンドによるIssueの隣接カラムへの移動（5.1）
- フィルタ適用後のIssue一覧の再描画
- ターミナルサイズ変更時のレスポンシブ対応（カラム数に応じた幅分配）
- カラム表示・非表示の切り替え（11.1, 11.2）。`hiddenCols`マップで管理し、表示カラムのみの仮想インデックスで`activeCol`を制御
- 非表示カラム数をステータスバー情報として返却（11.3）
- 最低1カラムの表示を維持するガード条件（11.4）

**Dependencies**
- Inbound: App — 画面遷移で表示 (P0)
- Outbound: IssueService — Issue一覧取得 (P0)
- Outbound: ProjectService — Projectステータス情報取得・ステータス移動 (P0)
- Outbound: Filter — フィルタ条件の適用 (P1)

**Contracts**: State [x]

##### State Management

```go
type BoardModel struct {
    columns      []StatusColumn
    activeCol    int
    cursorIndex  map[int]int
    scrollOffset map[int]int
    hiddenCols   map[int]bool
    loading      bool
    filterState  FilterState
    projectInfo  ProjectInfo
    width        int
    height       int
}

type StatusColumn struct {
    OptionID string
    Name     string
    Items    []ProjectItem
}

type ProjectItem struct {
    ItemID   string
    Issue    Issue
    StatusID string
}
```

**Implementation Notes**
- カラム非表示: `hiddenCols[colIndex] = true`で管理。描画時は`hiddenCols`をスキップし、表示カラムのみでレイアウト幅を分配
- カーソル移動: `activeCol`は常に表示カラムの仮想インデックスを参照。`visibleColumns()`ヘルパーで実カラムインデックスと仮想インデックスを変換
- 全復元: `hiddenCols`をクリアし、全カラムを表示に戻す
- ガード: `len(columns) - len(hiddenCols) <= 1`の場合、非表示操作を拒否しステータスメッセージで通知

#### Detail Model

| Field | Detail |
|-------|--------|
| Intent | Issue詳細情報の表示とプロパティのインライン編集 |
| Requirements | 4.1, 4.2, 4.3, 4.4, 5.4, 5.5, 5.6, 7.1, 7.2, 12.1, 12.2, 12.3, 12.4 |

**Responsibilities & Constraints**
- Issue本文のMarkdownレンダリング（glamour使用）
- メタデータ（ラベル、アサイニー、マイルストーン、日時）の表示
- コメント一覧の表示とスクロール
- ラベル/アサイニー/マイルストーンのインライン選択UI表示（12.1-12.3）
- 選択確定時にIssueServiceを通じてAPIに反映し、詳細画面を更新（12.4）
- 編集キャンセル時（Esc）は元のIssue詳細表示に戻る

**Dependencies**
- Inbound: App — Board上でIssue選択時に遷移 (P0)
- Outbound: IssueService — プロパティ更新 (P0)
- Outbound: RepoService — ラベル/コラボレーター/マイルストーン一覧取得 (P1)
- Outbound: Editor — テキスト編集への遷移 (P1)

**Contracts**: State [x]

##### State Management

```go
type DetailModel struct {
    issue         Issue
    comments      []Comment
    viewport      viewport.Model
    editingField  EditingField
    selectModel   SelectModel
    loading       bool
    errorMsg      string
    repoService   RepoService
    issueService  IssueService
}

type EditingField int

const (
    EditingNone EditingField = iota
    EditingLabels
    EditingAssignees
    EditingMilestone
)

type SelectModel struct {
    items     []SelectItem
    cursor    int
    selected  map[int]bool
    multiSelect bool
}

type SelectItem struct {
    ID       string
    Name     string
    Selected bool
}
```

**Implementation Notes**
- インライン編集: `l`/`a`/`m`キー押下で`editingField`を遷移。RepoServiceからメタデータを非同期取得し、`selectModel`に選択肢をセット
- 選択UI: `SelectModel`はラベル（マルチセレクト）、アサイニー（マルチセレクト）、マイルストーン（シングルセレクト）に対応。`multiSelect`フラグで切り替え
- 確定: Enterキーで選択を確定し、IssueService.UpdateIssueを呼び出し。成功時はIssueデータを更新して詳細表示に復帰
- キャンセル: Escキーで`editingField = EditingNone`に戻り、選択UIを閉じる
- メタデータキャッシュ: セッション中はRepoServiceの結果をキャッシュし、同一セッション内の再取得を回避

#### Editor Model

| Field | Detail |
|-------|--------|
| Intent | テキスト入力UIと外部エディタ（$EDITOR）の起動を管理する |
| Requirements | 5.2, 5.3, 6.1, 6.2, 6.3, 6.4, 7.1, 7.3 |

**Responsibilities & Constraints**
- タイトル等の短いテキストはインライン入力（bubbles/textinput）
- 本文・コメント等の長文は$EDITORでの外部エディタ起動
- 外部エディタ起動中はBubble Teaのtea.ExecCmd機能を使用
- 編集失敗時のテキスト保持

**Dependencies**
- Inbound: App, Detail — 編集操作のトリガー (P0)
- Outbound: IssueService — 変更のAPI反映 (P0)

**Contracts**: State [x]

##### State Management

```go
type EditorMode int

const (
    EditorInline EditorMode = iota
    EditorExternal
)

type EditorModel struct {
    mode         EditorMode
    textInput    textinput.Model
    content      string
    issueNumber  int
    editTarget   EditTarget
    creating     bool
    savedContent string
}

type EditTarget int

const (
    EditTitle EditTarget = iota
    EditBody
    EditComment
)
```

#### Filter Model

| Field | Detail |
|-------|--------|
| Intent | フィルタ条件・ソート条件のUI選択と適用 |
| Requirements | 3.1, 3.2, 3.3 |

**Responsibilities & Constraints**
- ラベル、アサイニー、マイルストーンのマルチセレクトUI
- ソート条件（作成日/更新日/コメント数）の選択
- 適用中のフィルタ条件の視覚化

**Dependencies**
- Inbound: App — フィルタパネル表示 (P0)
- Outbound: RepoService — ラベル/アサイニー/マイルストーン一覧取得 (P1)

**Contracts**: State [x]

##### State Management

```go
type SortField int

const (
    SortCreated SortField = iota
    SortUpdated
    SortComments
)

type SortDirection int

const (
    SortDesc SortDirection = iota
    SortAsc
)

type FilterState struct {
    labels      []string
    assignees   []string
    milestone   string
    sortField   SortField
    sortDir     SortDirection
}

type FilterModel struct {
    state        FilterState
    availLabels  []Label
    availUsers   []User
    availMilestones []Milestone
    activePane   FilterPane
    cursor       int
}
```

#### Help Model

| Field | Detail |
|-------|--------|
| Intent | キーバインドヘルプ表示 |
| Requirements | 8.3 |

**Contracts**: State [x]

**Implementation Notes**
- カラム表示・非表示のキーバインド（`d`/`D`）をヘルプ一覧に追加
- Issue詳細画面のインライン編集キー（`l`/`a`/`m`）をヘルプ一覧に追加

### CLI Layer

#### CLI Prompt - Project Selection

| Field | Detail |
|-------|--------|
| Intent | TUI起動前にCLIのインラインプロンプトでProject紐付け確認と選択を行う |
| Requirements | 10.1, 10.2, 10.3, 10.7, 10.8 |

**Responsibilities & Constraints**
- TUI（Bubble Tea）起動前にターミナル上でインラインプロンプトを表示（10.1）
- `charmbracelet/huh`の`Confirm`コンポーネントで「Bind a GitHub Project?」のYes/No確認を表示（10.1）
- Yes選択時にProjectService経由でリポジトリのProject一覧を取得し、`huh.Select`で選択肢を表示（10.2）
- No選択時はProject紐付けなしの状態でTUIを起動（10.3）
- 選択確定時にConfigService経由で設定ファイルに保存（10.4）
- `--config`フラグによる再設定時も同じフローを使用（10.7）
- Project無効時（削除済み等）はエラーメッセージを表示し、同じプロンプトフローにフォールバック（10.8）
- プロンプトはmain.go内の関数として実装し、Bubble Tea TUIとは完全に分離する
- `gh repo create`、`gh issue create`等のGitHub CLIと同様のUXを提供する

**Dependencies**
- Inbound: Main — 初回起動時または--config時に呼び出し (P0)
- Outbound: ProjectService — Project一覧取得 (P0)
- Outbound: ConfigService — 設定保存 (P0)
- External: charmbracelet/huh — インラインプロンプトUI (P0)

**Contracts**: Service [x]

##### Service Interface

```go
// promptProjectSelection はTUI起動前にCLIプロンプトでProject選択を実行する。
// 戻り値: 選択されたProject番号（0はProject紐付けなし）
func promptProjectSelection(
    projectSvc *service.ProjectService,
    repoRoot string,
) (int, error)
```

- Step 1: `huh.NewConfirm().Title("Bind a GitHub Project?")` でYes/No確認
- Step 2 (Yesの場合): `ProjectService.ListProjects()` でProject一覧取得
- Step 3: `huh.NewSelect().Title("Select a project")` でProject選択
- Step 4: `config.Save(repoRoot, Config{ProjectNumber: selected})` で設定保存
- Noの場合: 0を返し、Project紐付けなしでTUIを起動

**Implementation Notes**
- `huh`はCharm ecosystemの一部であり、Bubble Tea/Lip Glossと一貫したスタイリングを提供
- プロンプトはaltscreenを使用しないため、選択結果がターミナル履歴に残る（CLIツールの標準的な挙動）
- ローディング表示: Project一覧取得中は`huh`のSpinnerまたはシンプルなfmt.Printで「Fetching projects...」を表示

### Domain Layer

#### Issue Service

| Field | Detail |
|-------|--------|
| Intent | Issue CRUDおよび関連データ操作をUI層に提供する |
| Requirements | 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 6.3, 7.2, 9.1, 9.3, 12.1, 12.2, 12.3 |

**Responsibilities & Constraints**
- Issue一覧取得（GraphQL、ペジネーション付き）
- Issue作成・更新（タイトル、本文、ラベル、アサイニー、マイルストーン）
- コメント取得・追加
- API失敗時のエラー返却（ロールバックはUI層の責務）
- ステータス変更はProjectServiceの責務（Projects V2 APIを使用）

**Dependencies**
- Inbound: Board, Detail, Editor — CRUD操作 (P0)
- Outbound: GHClient — API通信 (P0)

**Contracts**: Service [x]

##### Service Interface

```go
type IssueService interface {
    ListIssues(ctx context.Context, opts ListIssuesOptions) ([]Issue, PageInfo, error)
    GetIssue(ctx context.Context, number int) (Issue, error)
    CreateIssue(ctx context.Context, input CreateIssueInput) (Issue, error)
    UpdateIssue(ctx context.Context, number int, input UpdateIssueInput) (Issue, error)
    AddComment(ctx context.Context, number int, body string) (Comment, error)
    ListComments(ctx context.Context, number int) ([]Comment, error)
}

type ListIssuesOptions struct {
    State     IssueState
    Labels    []string
    Assignee  string
    Milestone string
    Sort      SortField
    Direction SortDirection
    PerPage   int
    After     string
}

type CreateIssueInput struct {
    Title     string
    Body      string
    Labels    []string
    Assignees []string
    Milestone string
}

type UpdateIssueInput struct {
    Title     *string
    Body      *string
    Labels    *[]string
    Assignees *[]string
    Milestone *string
}
```

- Preconditions: ownerとrepoが解決済みであること
- Postconditions: API操作の結果が返却される。エラー時はerrorが非nil
- Invariants: ネットワーク障害時はラップされたエラーを返す。nilポインタはフィールド未変更を意味する

#### Project Service

| Field | Detail |
|-------|--------|
| Intent | GitHub Projects V2のステータスフィールド情報の取得およびIssueのステータス変更を管理する |
| Requirements | 2.2, 5.1, 10.2 |

**Responsibilities & Constraints**
- リポジトリに紐づくProjectV2の一覧取得・選択
- Statusフィールド（ProjectV2SingleSelectField）のoptions取得
- ProjectV2Itemの一覧取得（IssueとProjectV2Item IDのマッピング）
- Issueのステータス変更（updateProjectV2ItemFieldValueミューテーション）
- ProjectV2Item IDとIssue IDは異なるため、型レベルで区別する

**Dependencies**
- Inbound: Board — ステータスカラム情報取得・ステータス移動 (P0)
- Inbound: CLIPrompt — Project一覧取得 (P0)
- Outbound: GHClient — API通信 (P0)

**Contracts**: Service [x]

##### Service Interface

```go
type ProjectService interface {
    ListProjects(ctx context.Context, owner, repo string) ([]ProjectSummary, error)
    GetProjectFields(ctx context.Context, projectID string) (ProjectInfo, error)
    GetProjectItems(ctx context.Context, projectID string, cursor string) ([]ProjectItem, PageInfo, error)
    MoveItemStatus(ctx context.Context, projectID, itemID, fieldID, optionID string) error
    AddItemToProject(ctx context.Context, projectID, contentID string) (string, error)
}

type ProjectSummary struct {
    ID     string
    Number int
    Title  string
}

type ProjectInfo struct {
    ID          string
    Title       string
    StatusField StatusField
}

type StatusField struct {
    ID      string
    Name    string
    Options []StatusOption
}

type StatusOption struct {
    ID   string
    Name string
}
```

- Preconditions: ownerとrepoが解決済みであること。ProjectV2が存在すること
- Postconditions: ステータス変更はGitHub Projects V2に即座に反映される
- Invariants: MoveItemStatusにはProjectV2Item IDを使用する（Issue IDではない）。StatusFieldが存在しないProjectの場合はエラーを返す

#### Repo Service

| Field | Detail |
|-------|--------|
| Intent | リポジトリ情報とメタデータ（ラベル、コラボレーター、マイルストーン）の取得 |
| Requirements | 1.5, 3.1, 12.1, 12.2, 12.3 |

**Responsibilities & Constraints**
- カレントディレクトリからのリポジトリowner/name自動検出
- リポジトリのラベル・コラボレーター・マイルストーン一覧取得
- メタデータはフィルタUI、プロパティ編集UI、Issue詳細インライン編集UIで使用

**Dependencies**
- Inbound: Filter, Detail — メタデータ取得 (P1)
- Outbound: GHClient — API通信 (P0)

**Contracts**: Service [x]

##### Service Interface

```go
type RepoService interface {
    ResolveRepo() (RepoInfo, error)
    ListLabels(ctx context.Context) ([]Label, error)
    ListCollaborators(ctx context.Context) ([]User, error)
    ListMilestones(ctx context.Context) ([]Milestone, error)
}

type RepoInfo struct {
    Owner string
    Name  string
    Host  string
}
```

#### Config Service

| Field | Detail |
|-------|--------|
| Intent | リポジトリ単位の設定ファイル（`.gh-tuissue.json`）の読み書きを管理する |
| Requirements | 10.4, 10.5, 10.6, 10.7, 10.8 |

**Responsibilities & Constraints**
- リポジトリルートの`.gh-tuissue.json`の読み込み・書き込み
- 設定ファイルが存在しない場合はnilを返す（初回起動の判定に使用）
- JSONパースエラー時はデフォルト値にフォールバックし、警告を返す
- `--project`フラグの値は設定ファイルより優先されるが、ConfigServiceは設定ファイルの上書きを行わない（10.6の制約はmain.goで制御）

**Dependencies**
- Inbound: Main — 起動時の設定読み込み (P0)
- Inbound: CLIPrompt — 設定保存 (P0)
- Outbound: ConfigFile — ファイルI/O (P0)

**Contracts**: Service [x]

##### Service Interface

```go
type ConfigService interface {
    LoadConfig(repoRoot string) (*Config, error)
    SaveConfig(repoRoot string, config Config) error
}

type Config struct {
    ProjectNumber int `json:"project_number,omitempty"`
}
```

- Preconditions: repoRootが有効なディレクトリパスであること
- Postconditions: SaveConfig完了後、設定ファイルがディスクに書き込まれる
- Invariants: LoadConfigは設定ファイルが存在しない場合nil, nilを返す。JSONパースエラー時はnil, errorを返す

### Infrastructure Layer

#### GitHub API Client

| Field | Detail |
|-------|--------|
| Intent | go-gh v2を利用したGitHub REST/GraphQL APIへの認証済みアクセス |
| Requirements | 1.3, 1.4 |

**Responsibilities & Constraints**
- go-gh v2の`api.DefaultGraphQLClient()`と`api.DefaultRESTClient()`をラップ
- GraphQLクエリのビルドと実行
- RESTリクエストの実行
- 認証エラーの検出とユーザーフレンドリーなエラーメッセージ生成

**Dependencies**
- Inbound: IssueService, ProjectService, RepoService — API実行 (P0)
- External: go-gh v2 — 認証・HTTPクライアント (P0)
- External: GitHub API — REST v3 / GraphQL v4 (P0)

**Contracts**: Service [x]

##### Service Interface

```go
type GitHubClient interface {
    QueryGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error
    RESTGet(ctx context.Context, path string, result interface{}) error
    RESTPatch(ctx context.Context, path string, body interface{}, result interface{}) error
    RESTPost(ctx context.Context, path string, body interface{}, result interface{}) error
}
```

- Preconditions: gh auth loginが完了していること
- Postconditions: API応答がresultにデシリアライズされる
- Invariants: 未認証時はActionableなエラーメッセージ（`gh auth login`の実行を促す）を返す

## Data Models

### Domain Model

```mermaid
erDiagram
    ProjectV2 {
        string ID
        int Number
        string Title
    }
    StatusField {
        string ID
        string Name
    }
    StatusOption {
        string ID
        string Name
    }
    ProjectV2Item {
        string ItemID
        string StatusID
    }
    Issue {
        int Number
        string Title
        string Body
        IssueState State
        string NodeID
        string URL
        datetime CreatedAt
        datetime UpdatedAt
    }
    Label {
        string Name
        string Color
        string Description
    }
    User {
        string Login
        string Name
    }
    Milestone {
        int Number
        string Title
        string State
        datetime DueOn
    }
    Comment {
        int ID
        string Body
        datetime CreatedAt
    }
    Config {
        int ProjectNumber
    }

    ProjectV2 ||--|| StatusField : has
    StatusField ||--o{ StatusOption : has
    ProjectV2 ||--o{ ProjectV2Item : contains
    ProjectV2Item }o--|| Issue : references
    ProjectV2Item }o--|| StatusOption : currentStatus
    Issue ||--o{ Label : has
    Issue ||--o{ User : assignedTo
    Issue }o--o| Milestone : belongsTo
    Issue ||--o{ Comment : has
    Comment }o--|| User : author
    Issue }o--|| User : author
```

**Entities and Types**:

```go
type IssueState string

const (
    IssueOpen   IssueState = "OPEN"
    IssueClosed IssueState = "CLOSED"
)

type Issue struct {
    Number    int
    NodeID    string
    Title     string
    Body      string
    State     IssueState
    URL       string
    Author    User
    Labels    []Label
    Assignees []User
    Milestone *Milestone
    Comments  []Comment
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Label struct {
    Name        string
    Color       string
    Description string
}

type User struct {
    Login string
    Name  string
}

type Milestone struct {
    Number int
    Title  string
    State  string
    DueOn  *time.Time
}

type Comment struct {
    ID        int
    Body      string
    Author    User
    CreatedAt time.Time
}

type PageInfo struct {
    HasNextPage bool
    EndCursor   string
}

type Config struct {
    ProjectNumber int `json:"project_number,omitempty"`
}
```

- **Invariants**: IssueのNumberはリポジトリ内で一意。StateはOPENまたはCLOSEDのみ。NodeIDはGraphQLのグローバルノードID。ConfigのProjectNumberが0の場合はProject未紐付けを意味する。

### Data Contracts & Integration

**GraphQL Query — Project情報・ステータスフィールド取得**:

```graphql
query GetProjectAndStatusField($owner: String!, $repo: String!, $projectNumber: Int!) {
  repository(owner: $owner, name: $repo) {
    projectV2(number: $projectNumber) {
      id
      title
      fields(first: 20) {
        nodes {
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}
```

**GraphQL Query — ProjectV2Items取得（ステータス付き）**:

```graphql
query GetProjectItems($projectId: ID!, $cursor: String) {
  node(id: $projectId) {
    ... on ProjectV2 {
      items(first: 50, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          fieldValues(first: 20) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                optionId
                field { ... on ProjectV2SingleSelectField { name } }
              }
            }
          }
          content {
            ... on Issue {
              id
              number
              title
              body
              state
              url
              createdAt
              updatedAt
              author { login }
              labels(first: 10) { nodes { name color description } }
              assignees(first: 10) { nodes { login name } }
              milestone { number title state dueOn }
            }
          }
        }
      }
    }
  }
}
```

**GraphQL Mutation — ステータス変更**:

```graphql
mutation UpdateItemStatus($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
  updateProjectV2ItemFieldValue(
    input: { projectId: $projectId, itemId: $itemId, fieldId: $fieldId, value: { singleSelectOptionId: $optionId } }
  ) {
    projectV2Item { id }
  }
}
```

**GraphQL Query — リポジトリのProject一覧取得**:

```graphql
query ListProjects($owner: String!, $repo: String!) {
  repository(owner: $owner, name: $repo) {
    projectsV2(first: 20) {
      nodes {
        id
        number
        title
      }
    }
  }
}
```

**GraphQL Query — Issue一覧取得（Project未使用時のフォールバック）**:

```graphql
query ListIssues($owner: String!, $name: String!, $first: Int!, $after: String, $states: [IssueState!]) {
  repository(owner: $owner, name: $name) {
    issues(first: $first, after: $after, states: $states, orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes {
        id
        number
        title
        body
        state
        url
        createdAt
        updatedAt
        author { login }
        labels(first: 10) { nodes { name color description } }
        assignees(first: 10) { nodes { login name } }
        milestone { number title state dueOn }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
}
```

**REST API — Issue更新**: `PATCH /repos/{owner}/{repo}/issues/{issue_number}`

**REST API — コメント追加**: `POST /repos/{owner}/{repo}/issues/{issue_number}/comments`

**設定ファイル — `.gh-tuissue.json`**:

```json
{
  "project_number": 1
}
```

## Error Handling

### Error Strategy
- GitHub APIエラーはドメインエラー型にラップしてUI層に返す
- UI層はエラーメッセージをステータスバーに表示し、操作前の状態を維持する
- 認証エラーは専用メッセージで`gh auth login`を促す
- 設定ファイルのエラーはフォールバック動作で対処する

### Error Categories and Responses

**認証エラー**: 未認証 / トークン期限切れ → `gh auth login`の実行を促すメッセージ表示

**ネットワークエラー**: タイムアウト / 接続失敗 → エラーメッセージ表示、直前のデータで表示を維持 (9.3)

**APIエラー (4xx)**: 権限不足 / リソース未発見 / バリデーション失敗 → 具体的なエラー内容をステータスバーに表示

**更新失敗**: Issue編集のAPI失敗 → エラーメッセージ表示、ローカル表示を変更前に復帰 (5.7)

**Projects V2エラー**: Projectが存在しない / Statusフィールドが未設定 / projectスコープ不足 → 具体的なエラーメッセージ表示。Projectが存在しない場合はOpen/Closedフォールバックを提示

**ステータス移動失敗**: updateProjectV2ItemFieldValueの失敗 → エラーメッセージ表示、Issueを元のカラムに維持

**設定ファイルエラー**: JSONパース失敗 / ファイルI/Oエラー → 警告メッセージ表示、デフォルト値（Project未紐付け）にフォールバック

**無効Project参照**: 設定ファイルのProject番号が無効（削除済み等） → エラーメッセージ表示、Project選択UIを再表示 (10.8)

**インライン編集失敗**: ラベル/アサイニー/マイルストーン更新のAPI失敗 → エラーメッセージ表示、Issue詳細の表示を変更前に復帰、選択UIを閉じる

```go
type AppError struct {
    Code    ErrorCode
    Message string
    Err     error
}

type ErrorCode int

const (
    ErrAuth ErrorCode = iota
    ErrNetwork
    ErrNotFound
    ErrPermission
    ErrValidation
    ErrRateLimit
    ErrProjectNotFound
    ErrStatusFieldMissing
    ErrConfigLoad
    ErrConfigSave
    ErrUnknown
)
```

## Testing Strategy

### Unit Tests
- IssueServiceの各メソッド（フィルタ適用、ソート、ペジネーション）
- ProjectServiceのステータスフィールド取得・ステータス移動ロジック
- FilterStateの適用ロジック
- RepoServiceのリポジトリ自動検出ロジック
- AppErrorのエラーコード分類ロジック
- BoardModelの動的カラム構築・隣接カラム解決ロジック
- BoardModelのカラム表示・非表示ロジック（hiddenCols管理、仮想インデックス変換、最低1カラムガード）
- ConfigServiceの設定ファイル読み書き（存在しない場合、パースエラー、正常読み書き）
- DetailModelのインライン編集状態遷移（EditingNone→EditingLabels→確定/キャンセル→EditingNone）
- SelectModelの選択ロジック（マルチセレクト、シングルセレクト、プリセレクト）

### Integration Tests
- GHClientを通じた実際のGraphQLクエリ実行（テストリポジトリ使用）
- Projects V2 APIを通じたステータスフィールド取得・ステータス移動
- Issue作成→更新のライフサイクルフロー
- 外部エディタ起動・内容取得のフロー
- ConfigServiceの設定ファイル永続化（書き込み→読み込みの往復テスト）
- DetailModelからのインライン編集→API更新→画面反映フロー

### E2E/UI Tests
- teatestライブラリを使用したBubble Teaモデルのテスト
- カンバン表示→Issue選択→詳細表示の画面遷移フロー
- キーバインド操作（h/j/k/l移動、矢印キー移動、Enter選択、Esc戻る）
- ステータス移動コマンドによるIssueのカラム間移動
- 編集完了後のカンバンボード画面への復帰確認
- altscreenモードによる終了時の画面クリア確認
- 初回起動→Project選択→設定保存→次回起動の自動読み込みフロー
- カラム非表示→レイアウト再調整→全復元フロー
- Issue詳細画面→lキー→ラベル選択→確定→API反映→画面更新フロー

### Performance
- 100件以上のIssueのペジネーション付き一覧取得
- N個のカスタムカラムでのレイアウト計算
- ターミナルリサイズ時のレスポンシブ再計算
