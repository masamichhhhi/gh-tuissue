# Research & Design Decisions

## Summary

- **Feature**: `gh-issue-kanban-tui`
- **Discovery Scope**: New Feature（グリーンフィールド）→ Extension（GitHub Projects V2統合の追加）
- **Key Findings**:
  - Bubble Tea v2がElmアーキテクチャベースの成熟したGoTUIフレームワークとして最適
  - go-gh v2ライブラリによりGitHub CLI認証・APIアクセスをシームレスに統合可能
  - GraphQL APIが複合データ取得（Issue+ラベル+アサイニー+マイルストーン）に最適
  - GitHub Projects V2 GraphQL APIにより、カスタムステータス（SingleSelectField）ベースのカンバンカラム構成が可能
  - ステータス変更は`updateProjectV2ItemFieldValue`ミューテーションで実現。ProjectV2Item IDとField/Option IDの3つが必要
  - 設定ファイル（`.gh-tuissue.json`）によるProject紐付けの永続化で、起動時の手間を最小化
  - カラム表示・非表示の切り替えはBoardModelの状態管理で実現可能、Bubble Teaのイベントループ内で完結
  - Issue詳細画面からのインライン編集は既存のFilter UIコンポーネントの再利用パターンで効率的に実装可能

## Research Log

### Go TUIフレームワーク選定

- **Context**: カンバンボード表示に適したGo TUIフレームワークの調査
- **Sources Consulted**: GitHub charmbracelet/bubbletea, charmbracelet/lipgloss, charmbracelet/bubbles
- **Findings**:
  - Bubble Tea v2（2026年2月リリース）: Elmアーキテクチャ（Model→Update→View）、Cursed Rendererで30%高速化、Mode 2026対応
  - インポートパス: `charm.land/bubbletea/v2`
  - Lip Gloss: CSS風ターミナルスタイリング、カラーダウンサンプリング対応
  - Bubbles: テーブル、リスト、テキスト入力、ビューポート等のコンポーネントライブラリ
  - 40.3k GitHub stars、NVIDIA/Microsoft/AWS/GitHub等で本番利用
  - カンバンボードの実装例が複数存在（igormomc/Golang-KanbanBoard等）
- **Implications**: Bubble Tea v2 + Lip Gloss + Bubblesの組み合わせが最適。既存のカンバン実装パターンを参考にできる

### GitHub CLI拡張機能開発

- **Context**: gh extensionとしてのアーキテクチャとAPI統合方法の調査
- **Sources Consulted**: GitHub Docs, go-gh v2 pkg.go.dev, cli/gh-extension-precompile
- **Findings**:
  - `gh extension create --precompiled=go`でプロジェクト初期化
  - go-gh v2（`github.com/cli/go-gh/v2`）がGitHub CLI統合の公式ライブラリ
  - `api.DefaultRESTClient()` / `api.DefaultGraphQLClient()` で認証済みAPIクライアント取得
  - 認証は自動的にGH_TOKEN/gh auth tokenを利用、手動管理不要
  - リリース自動化: cli/gh-extension-precompile GitHub Action
  - バイナリ命名規則: `gh-tuissue-<os>-<arch>[.exe]`
- **Implications**: go-ghライブラリで認証とAPIアクセスを完全に抽象化できるため、認証実装のコストが最小

### GitHub Issues API選定

- **Context**: IssueデータのCRUD操作に最適なAPI方式の調査
- **Sources Consulted**: GitHub REST API Docs, GitHub GraphQL API Docs
- **Findings**:
  - REST API: 5,000 req/hr、ページベースペジネーション（per_page最大100）
  - GraphQL API: 2,000 points/min、カーソルベースペジネーション
  - GraphQLなら1リクエストでIssue+ラベル+アサイニー+マイルストーン+コメントを取得可能
  - REST APIでは同等のデータ取得にN+1問題が発生
  - ミューテーション: createIssue, updateIssue, closeIssue, reopenIssue, addComment等
  - カーソルベースペジネーションはリアルタイムデータに対してより信頼性が高い
- **Implications**: 読み取り操作はGraphQL、一部の更新操作はRESTのハイブリッドアプローチが最適

### GitHub Projects V2 GraphQL API

- **Context**: カンバンカラムをOpen/Closedの2分割ではなく、GitHub Projectで定義されたカスタムステータス（Column）に対応する要件変更に伴う調査
- **Sources Consulted**: GitHub GraphQL API Docs（Projects V2）、GitHub公式ドキュメント
- **Findings**:
  - GitHub Projects V2はGraphQL APIで操作可能。主要な型は`ProjectV2`、`ProjectV2Item`、`ProjectV2SingleSelectField`
  - ステータスは`ProjectV2SingleSelectField`の`options`として定義される（例: Todo, In Progress, Done）
  - 各optionは固有の`id`と`name`を持つ
  - Issueのステータス取得: `ProjectV2Item.fieldValues`から`ProjectV2ItemFieldSingleSelectValue`を取得
  - ステータス更新: `updateProjectV2ItemFieldValue`ミューテーションで`projectId`、`itemId`（ProjectV2Item ID、Issue IDとは異なる）、`fieldId`、`singleSelectOptionId`が必要
  - IssueをProjectに追加: `addProjectV2ItemById`ミューテーション（冪等、既存の場合は既存itemを返す）
  - リポジトリの全Projects取得: `repository.projectsV2(first: N)`
  - 特定Project取得: `repository.projectV2(number: N)`
  - "Status"フィールド名はユーザーがカスタマイズ可能なため、名前ベースで検索する必要がある
  - go-gh v2では`client.Do(queryString, variables, &response)`で生のGraphQLクエリを実行可能（union型やinline fragmentが多いProjects V2クエリに適切）
- **Implications**:
  - 新たに`ProjectService`コンポーネントが必要（Project/Field/Option情報の取得・キャッシュ）
  - Board Modelは動的なカラム数に対応する必要がある（固定2カラム→N個のカスタムカラム）
  - IssueとProjectV2Itemの関連付け管理が必要（Issue IDとItem IDのマッピング）
  - `--project`フラグまたは自動検出でProjectを選択するUI/CLIフローが必要
  - ステータス移動は隣接カラムへの左右移動コマンドで実装（option配列のインデックス操作）

### TUI終了時の画面クリアとBubble Tea v2

- **Context**: TUI終了後にターミナルに残像が残るバグの修正方法調査
- **Sources Consulted**: Bubble Tea v2ドキュメント
- **Findings**:
  - Bubble Tea v2では`tea.Program`にaltscreenモードを使用可能（`tea.WithAltScreen()`オプション）
  - altscreenモードを使用すると、終了時に自動的に元のターミナル画面に復帰し、TUIの表示が残らない
  - `p.Quit()`または`tea.Quit`コマンドで正常終了時にaltscreenが解除される
- **Implications**: `tea.WithAltScreen()`オプションを使用するだけで終了時の画面クリアが実現可能

### プロジェクト紐付け永続化の設計調査

- **Context**: Requirement 10で要求されるProject紐付けの永続化方式の調査
- **Sources Consulted**: 既存のgh extension設定パターン、Go標準ライブラリ（encoding/json, os）
- **Findings**:
  - gh extensionの設定ファイルはリポジトリルートに配置するのが一般的（`.gh-tuissue.json`）
  - Go標準ライブラリの`encoding/json`と`os`パッケージで十分実装可能（外部依存不要）
  - 設定ファイルの検出はgitリポジトリルートからの相対パスで行う（`git rev-parse --show-toplevel`相当）
  - `--project`フラグは設定ファイルより優先するが、設定ファイルを上書きしない（10.6）
  - `--config`フラグで設定変更UIを再表示する仕組みが必要（10.7）
  - Projectが削除された場合のバリデーションには、起動時のProject存在確認APIコールが必要（10.8）
- **Implications**:
  - 新たに`ConfigService`コンポーネントが必要（設定ファイルの読み書き）
  - 初回起動フローにProject選択UIを挟む（Yes/No → Project一覧選択）
  - main.goの起動フローに設定ファイル読み込み→バリデーション→フォールバックのロジックが必要

### Project選択UIのCLIプロンプト移行

- **Context**: Requirement 10のProject選択UIをBubble Tea TUI内の画面からTUI起動前のCLIインラインプロンプトに変更
- **Sources Consulted**: charmbracelet/huh、gh CLI（gh repo create, gh issue create）のUXパターン
- **Findings**:
  - `charmbracelet/huh`はCharm ecosystemの公式フォーム/プロンプトライブラリ
  - `huh.NewConfirm()`でYes/No確認、`huh.NewSelect()`でリスト選択を提供
  - Lip Glossと統合済みのスタイリングで、既存のCharm依存と一貫したルック&フィール
  - `gh` CLI本体やその他のモダンCLIツール（npm init, cargo init等）と同様のインラインプロンプトUX
  - TUI（altscreen）起動前に実行するため、選択結果がターミナル履歴に残る
  - Bubble Tea TUIの画面遷移（ViewProjectSelect）が不要になり、AppModelが簡素化
  - ProjectSelectModel（internal/ui/project_select.go）は廃止し、main.go内の関数に置き換え
- **Implications**:
  - `charmbracelet/huh`を新規依存として追加
  - ProjectSelectModelの削除、AppModelからViewProjectSelect状態の削除
  - main.goにProject選択ロジックを移動（TUI起動前に完結）
  - ProjectServiceはCLI層から直接呼び出す（UI層を経由しない）

### カラム表示・非表示の設計調査

- **Context**: Requirement 11で要求されるカラム表示・非表示切り替え機能の調査
- **Sources Consulted**: 既存のboard.go実装、Bubble Teaのイベント処理パターン
- **Findings**:
  - 既存の`BoardModel`は`columns []StatusColumn`と`activeCol int`で管理
  - カラム非表示は`hiddenCols map[int]bool`でインデックスベースの管理が最もシンプル
  - 表示カラムのみの「仮想インデックス」を導入し、`activeCol`は常に表示カラムを参照するよう変換
  - 非表示カラム数はステータスバーに表示（既存の`statusMsg`を拡張）
  - 少なくとも1カラムの表示を強制するガード条件が必要（11.4）
  - レイアウト再計算は既存の幅分配ロジックを表示カラム数ベースに変更するだけ
- **Implications**:
  - BoardModelに`hiddenCols`フィールドと可視カラム操作ヘルパーを追加
  - キーバインドの追加（`d`で非表示、`D`で全復元等）
  - ステータスバーへの非表示カラム数表示

### Issue詳細画面インライン編集の設計調査

- **Context**: Requirement 12で要求されるIssue詳細画面からのプロパティインライン編集の調査
- **Sources Consulted**: 既存のdetail.go、filter.go、app.goの実装
- **Findings**:
  - 既存のdetail.goに`editType`列挙型（`editLabels`, `editAssignees`, `editMilestone`）とキーハンドラ（l/a/mキー）が定義済み
  - 現状はプレースホルダメッセージ（"use filter panel (not yet inline)"）を表示するのみ
  - 既存のfilter.goにマルチセレクトUIの実装あり（ラベル/アサイニー/マイルストーン選択）
  - filter.goの選択UIロジックを再利用可能な`SelectModel`として抽出し、Detail画面に埋め込むアプローチが最適
  - Detail画面のUpdate()内で選択UIのイベント処理をデリゲートし、完了時にAPIコールを実行
  - RepoServiceからメタデータ（ラベル/コラボレーター/マイルストーン一覧）を取得する必要あり
- **Implications**:
  - 既存の選択UIパターンを再利用できるため、新規コンポーネントの作成は最小限
  - DetailModelに`selectModel`フィールドと`editingField`状態を追加
  - RepoServiceへの依存をDetailModelに追加

### ステータス移動の楽観的UI更新（Optimistic UI Update）

- **Context**: H/Lキーによるステータス移動時のAPI応答待ちによるUX低下への対応策調査
- **Sources Consulted**: Bubble Tea v2のtea.Cmdパターン、既存のhandleStatusMove実装（app.go）、Elm Architecture における楽観的更新パターン
- **Findings**:
  - 現状の実装はAPI呼び出し（MoveItemStatus）→ 成功後にreloadBoardData()で全件リロードの2段階。ネットワーク往復が2回発生し、その間UIがブロックされる
  - Bubble Teaのtea.Cmdは非同期メッセージを返す仕組みであり、API呼び出し前にModel状態を先行更新することで楽観的UIが実現可能
  - 楽観的UI更新のパターン: (1) ローカルデータを即座に更新 → (2) tea.Cmdで非同期API呼び出し → (3) 成功時はそのまま維持 → (4) 失敗時はロールバック
  - ロールバックの実装: 変更前のStatusIDと元のカラムインデックスを`statusMoveMsg`に含めて返し、失敗時にBoardModel内のアイテムを元の状態に復元する
  - reloadBoardData()の省略: API成功時は楽観的更新済みのローカルデータがすでに正しい状態なので、全件リロードは不要。他ユーザーの変更は`r`（リフレッシュ）キーで手動同期
  - BoardModel.items内のProjectItemのStatusIDとカラム配置を直接書き換える関数（MoveItemToColumn等）が必要
- **Implications**:
  - handleStatusMove()のリファクタリング: ローカル更新を先に実行し、tea.CmdでAPI呼び出しを返す
  - statusMoveMsgにロールバック情報（元のStatusID、元のカラムインデックス）を追加
  - statusMoveMsg受信時: 成功→何もしない（reloadBoardData不要）、失敗→ロールバック+エラー表示
  - BoardModelにアイテムのカラム間移動メソッドを追加

## Architecture Pattern Evaluation


| Option                        | Description                  | Strengths                     | Risks / Limitations  | Notes                   |
| ----------------------------- | ---------------------------- | ----------------------------- | -------------------- | ----------------------- |
| Elm Architecture (Bubble Tea) | Model-Update-View一方向データフロー   | 状態管理の明確さ、テスタビリティ、Bubble Tea標準 | 複雑なネスト状態の管理が煩雑になる可能性 | Bubble Tea v2の標準パターンで採用 |
| レイヤードアーキテクチャ                  | UI/Domain/Infrastructure層の分離 | 関心の分離、テスタビリティ                 | 小規模プロジェクトにはオーバーヘッド   | GitHub API層とTUI層の分離に適用  |


## Design Decisions

### Decision: TUIフレームワーク — Bubble Tea v2

- **Context**: カンバンボードUIの描画・イベント処理フレームワーク選定
- **Alternatives Considered**:
  1. Bubble Tea v2 — Elmアーキテクチャ、最も活発なエコシステム
  2. tview — tcellベース、フォーム・テーブル等の豊富なウィジェット
  3. tcell — 低レベルAPI、柔軟だが冗長
- **Selected Approach**: Bubble Tea v2
- **Rationale**: Go TUI分野で最大のコミュニティ、Lip Gloss/Bubblesとの統合、カンバン実装の先行例あり
- **Trade-offs**: v2は新しいためv1からのマイグレーション情報が限定的だが、グリーンフィールドなので問題なし
- **Follow-up**: v2のAPI変更（View()がtea.View構造体を返す）に注意

### Decision: API方式 — GraphQL優先 + REST補完

- **Context**: GitHub Issues APIのアクセス方式選定
- **Alternatives Considered**:
  1. REST API only — シンプルだがN+1問題
  2. GraphQL only — 効率的だがミューテーション操作のエラーハンドリングが複雑
  3. GraphQL読み取り + REST更新 — 各APIの強みを活用
- **Selected Approach**: GraphQL優先（読み取り）、REST補完（一部更新操作）
- **Rationale**: カンバン表示では1リクエストでIssue+関連データを効率的に取得する必要がある
- **Trade-offs**: 2種類のAPIクライアント管理が必要だが、go-ghが両方を提供するため実装コストは低い
- **Follow-up**: GraphQLクエリの複雑さがrate limitポイントに与える影響を実装時に検証

### Decision: GitHub CLI統合 — go-gh v2

- **Context**: gh extension としての認証・API統合方式
- **Selected Approach**: go-gh v2ライブラリによる完全統合
- **Rationale**: 認証の自動解決、REST/GraphQLクライアントの提供、gh公式サポート
- **Trade-offs**: go-ghへの依存だが、gh extensionとして公式推奨されているため問題なし

### Decision: カンバンカラム — GitHub Projects V2ステータスベース

- **Context**: カンバンカラムの構成をOpen/Closed 2分割からProject定義のカスタムステータスに変更
- **Alternatives Considered**:
  1. Open/Closed固定2カラム — シンプルだがProject Boardの柔軟性を活かせない
  2. ラベルベースのカスタムカラム — ラベルは本来の用途と異なる
  3. GitHub Projects V2のStatusフィールド連動 — Projectの設計意図に沿った自然な連携
- **Selected Approach**: GitHub Projects V2のStatusフィールド連動
- **Rationale**: GitHub Projectsはステータス管理の公式機能であり、ユーザーが既にProject Boardで定義したワークフローをTUIにそのまま反映できる
- **Trade-offs**: Projects V2 APIへの追加依存が発生し、GraphQLクエリの複雑性が増す。Projectが未設定のリポジトリにはフォールバック（Open/Closed）が必要
- **Follow-up**: `--project`フラグによるProject番号指定のUI設計。複数Projectが存在する場合の選択UI

### Decision: 設定ファイル形式 — JSON（`.gh-tuissue.json`）

- **Context**: Project紐付け設定の永続化形式の選定
- **Alternatives Considered**:
  1. JSON（`.gh-tuissue.json`） — Go標準ライブラリで対応、シンプル
  2. YAML — 人間にとって読みやすいが外部依存が必要
  3. TOML — Go標準ライブラリ非対応、外部依存
- **Selected Approach**: JSON形式（`.gh-tuissue.json`）
- **Rationale**: Go標準の`encoding/json`で完結。設定項目が少なくJSONで十分。`.gitignore`推奨（ユーザーごとに異なるProject紐付けの可能性）
- **Trade-offs**: YAMLほど人間に優しくないが、設定項目がProject番号程度なので問題なし
- **Follow-up**: 将来的に設定項目が増えた場合のマイグレーション方針

### Decision: Project選択UI — CLIインラインプロンプト（charmbracelet/huh）

- **Context**: Project選択UIの提供方式の再設計
- **Alternatives Considered**:
  1. Bubble Tea TUI内のProjectSelect画面 — 既存実装。フルスクリーンTUI内で画面遷移として管理
  2. CLIインラインプロンプト（huh） — TUI起動前にターミナル上でインラインプロンプト表示
  3. コマンドラインフラグのみ — `--project`フラグでの指定を必須化
- **Selected Approach**: CLIインラインプロンプト（charmbracelet/huh）
- **Rationale**: `gh repo create`等のGitHub CLIツールと一貫したUXを提供。Project選択はTUIの機能ではなく起動前の設定ステップであり、CLIプロンプトの方が適切。AppModelの複雑性が低減。altscreenを使わないため選択結果がターミナル履歴に残り、ユーザーが設定内容を確認しやすい
- **Trade-offs**: huhが新規依存として追加されるが、Charm ecosystemの一部であり既存依存との一貫性がある
- **Follow-up**: 既存のProjectSelectModel（internal/ui/project_select.go）の廃止

### Decision: インライン編集UI — 選択リストのDetail画面埋め込み

- **Context**: Issue詳細画面からのプロパティ編集UI方式の選定
- **Alternatives Considered**:
  1. Filter画面への遷移 — 既存実装をそのまま使えるが画面遷移が発生
  2. オーバーレイモーダル — 実装が複雑でBubble Teaのパターンに合わない
  3. Detail画面内に選択リストを埋め込み — 画面遷移なし、既存UIロジックの再利用可能
- **Selected Approach**: Detail画面内に選択リストを埋め込み
- **Rationale**: 画面遷移なしでスムーズな編集体験を提供。既存のfilter.goの選択UIパターンを再利用
- **Trade-offs**: DetailModelの状態管理が複雑化するが、`editingField`による明確な状態遷移で管理可能
- **Follow-up**: 選択リストの表示位置・サイズのレイアウト調整

## Risks & Mitigations

- Bubble Tea v2のAPI安定性 — v2は2026年2月リリースで新しいが、メジャーバージョンとして安定版扱い。コミュニティ採用も進行中
- GraphQL APIのrate limit — 複雑なクエリはポイント消費が多い。クエリの最適化とローカルキャッシュで対応
- 大量Issue時のパフォーマンス — ペジネーションとインクリメンタルロードで対応。初回表示は直近のOpen Issueに限定
- Projects V2のトークンスコープ — `project`スコープ（classic PAT）または`read:project`/`write:project`（fine-grained PAT）が必要。gh auth loginのデフォルトスコープに含まれない場合はユーザーに追加を促す
- ProjectV2Item IDとIssue IDの混同 — ミューテーションにはProjectV2Item IDが必要。Issue IDとの混同を防ぐために型レベルで区別する
- "Status"フィールドの不在 — Projectに"Status"フィールドが存在しない場合のフォールバック処理が必要
- 設定ファイルの破損・不正 — JSONパースエラー時はデフォルト値にフォールバックし、ユーザーに警告メッセージを表示
- 削除済みProjectの参照 — 起動時にProject存在確認APIコールを実行し、失敗時はProject選択UIを再表示（10.8）

## References

- [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) — Go TUIフレームワーク（v2、Elmアーキテクチャ）
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — ターミナルスタイリング
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUIコンポーネントライブラリ
- [go-gh v2](https://pkg.go.dev/github.com/cli/go-gh/v2) — GitHub CLI拡張機能ライブラリ
- [GitHub GraphQL API](https://docs.github.com/en/graphql) — Issue管理用GraphQL API
- [GitHub GraphQL API - Projects V2](https://docs.github.com/en/graphql/reference/objects#projectv2) — Projects V2 GraphQL API
- [GitHub REST API - Issues](https://docs.github.com/en/rest/issues) — Issue管理用REST API
- [gh-extension-precompile](https://github.com/cli/gh-extension-precompile) — リリース自動化Action
