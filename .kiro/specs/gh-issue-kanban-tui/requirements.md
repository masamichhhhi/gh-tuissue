# Requirements Document

## Introduction

GitHub CLI拡張機能（`gh extension`）として開発する、GitHub Issueをカンバンボード形式でステータスごとに一覧表示し、Issue内容やプロパティの編集を行えるTUIツール。`gh issue-tui`コマンドとして動作し、GitHub CLIの認証基盤を活用する。OSSとして公開予定。

## Requirements

### Requirement 1: GitHub CLI拡張機能としての動作

**Objective:** As a 開発者, I want `gh extension`として本ツールをインストール・実行したい, so that GitHub CLIのエコシステムに統合された形でIssue管理ができる

#### Acceptance Criteria

1. The gh-tuissue shall `gh extension install`でインストール可能なGitHub CLI拡張機能として動作する
2. The gh-tuissue shall `gh issue-tui`コマンドで起動する
3. The gh-tuissue shall GitHub CLIの認証トークン（`gh auth token`）を利用してGitHub APIにアクセスする
4. If GitHub CLIがインストールされていないまたは未認証の場合, the gh-tuissue shall エラーメッセージを表示し、`gh auth login`の実行を促す
5. When カレントディレクトリがGitリポジトリ内である時, the gh-tuissue shall リモートURLからリポジトリのowner/nameを自動検出する
6. When ユーザーが`--repo owner/name`フラグを指定した時, the gh-tuissue shall 指定されたリポジトリに接続する

### Requirement 2: Issueカンバンボード表示

**Objective:** As a 開発者, I want Issueをステータスごとのカンバンカラムで一覧表示したい, so that プロジェクトの進捗状況を視覚的に把握できる

#### Acceptance Criteria

1. When ツールが起動した時, the gh-tuissue shall リポジトリのIssueをカンバンボード形式で表示する
2. The gh-tuissue shall ユーザーがリポジトリのGitHub Projectで定義したステータス（Column）ごとにカラムを分けて表示する（Open/Closedの2分割ではなく、Project Boardのカスタムステータスを反映する）
3. The gh-tuissue shall 各Issueカードにタイトル、Issue番号、ラベル、アサイニーを表示する
4. When Issueの数がカラムの表示領域を超えた時, the gh-tuissue shall カラム内でスクロール可能にする
5. The gh-tuissue shall ターミナルのウィンドウサイズに応じてレイアウトを調整する

### Requirement 3: Issue一覧のフィルタリング・ソート

**Objective:** As a 開発者, I want Issueをラベル・アサイニー・マイルストーンなどでフィルタリングしたい, so that 必要なIssueを素早く見つけられる

#### Acceptance Criteria

1. When ユーザーがフィルタ操作を行った時, the gh-tuissue shall ラベル、アサイニー、マイルストーンによるフィルタリングを適用する
2. When ユーザーがソート操作を行った時, the gh-tuissue shall 作成日、更新日、コメント数によるソートを適用する
3. When フィルタが適用されている時, the gh-tuissue shall 現在のフィルタ条件を画面上に表示する
4. When フィルタ結果が0件の時, the gh-tuissue shall 該当Issueがない旨を表示する

### Requirement 4: Issue詳細閲覧

**Objective:** As a 開発者, I want カンバン上のIssueを選択して詳細内容を閲覧したい, so that ターミナルからIssueの全情報を確認できる

#### Acceptance Criteria

1. When ユーザーがカンバン上のIssueを選択した時, the gh-tuissue shall Issue詳細ビューを表示する
2. The gh-tuissue shall Issue詳細にタイトル、本文（Markdownレンダリング）、ラベル、アサイニー、マイルストーン、作成日、更新日を表示する
3. The gh-tuissue shall Issue詳細にコメント一覧を表示する
4. When Issue本文が長い場合, the gh-tuissue shall 詳細ビュー内でスクロール可能にする

### Requirement 5: Issueプロパティ編集

**Objective:** As a 開発者, I want TUI上からIssueのプロパティを編集したい, so that ブラウザを開かずにIssue管理を完結できる

#### Acceptance Criteria

1. When ユーザーがIssueのステータス変更を実行した時, the gh-tuissue shall IssueをProject Boardの隣接するステータス（カラム）に移動し、カンバン上の表示を更新する（左移動・右移動のコマンドで1つずつステータスを変更可能にする）
2. When ユーザーがIssueのタイトル編集を実行した時, the gh-tuissue shall インライン編集UIを表示し、変更をGitHubに反映する
3. When ユーザーがIssueの本文編集を実行した時, the gh-tuissue shall 外部エディタ（$EDITOR）を起動し、変更をGitHubに反映する
4. When ユーザーがラベルの追加・削除を実行した時, the gh-tuissue shall リポジトリの既存ラベル一覧から選択可能にし、変更をGitHubに反映する
5. When ユーザーがアサイニーの変更を実行した時, the gh-tuissue shall リポジトリのコラボレーター一覧から選択可能にし、変更をGitHubに反映する
6. When ユーザーがマイルストーンの変更を実行した時, the gh-tuissue shall リポジトリの既存マイルストーン一覧から選択可能にし、変更をGitHubに反映する
7. If GitHub APIへの更新リクエストが失敗した場合, the gh-tuissue shall エラーメッセージを表示し、ローカル表示を変更前の状態に戻す
8. When Issueのプロパティ編集が完了した時, the gh-tuissue shall 正しくIssue一覧（カンバンボード）画面に戻り、操作を継続可能にする

### Requirement 6: Issue新規作成

**Objective:** As a 開発者, I want TUI上から新しいIssueを作成したい, so that ターミナルワークフローの中でシームレスにIssueを追加できる

#### Acceptance Criteria

1. When ユーザーがIssue新規作成を実行した時, the gh-tuissue shall タイトル入力フォームを表示する
2. When ユーザーがIssue本文を入力する時, the gh-tuissue shall 外部エディタ（$EDITOR）を起動する
3. When Issueが正常に作成された時, the gh-tuissue shall カンバンボードに新しいIssueを追加表示する
4. Where ラベル・アサイニー・マイルストーンの指定が可能な場合, the gh-tuissue shall 作成時にこれらを設定可能にする

### Requirement 7: コメント操作

**Objective:** As a 開発者, I want Issue詳細画面からコメントを追加したい, so that ターミナルからIssueのディスカッションに参加できる

#### Acceptance Criteria

1. When ユーザーがコメント追加を実行した時, the gh-tuissue shall 外部エディタ（$EDITOR）を起動してコメント入力を受け付ける
2. When コメントが正常に投稿された時, the gh-tuissue shall コメント一覧を更新して新しいコメントを表示する
3. If コメント投稿が失敗した場合, the gh-tuissue shall エラーメッセージを表示し、入力内容を保持する

### Requirement 8: キーボードナビゲーション

**Objective:** As a 開発者, I want キーボードのみで全操作を完結したい, so that マウスなしで効率的にIssue管理ができる

#### Acceptance Criteria

1. The gh-tuissue shall Vimライクなキーバインド（h/j/k/l）および矢印キー（←/↓/↑/→）でカンバンカラム間・Issue間のカーソル移動をサポートする
2. The gh-tuissue shall Enterキーで選択、Escキーで戻る操作をサポートする
3. The gh-tuissue shall `?`キーでキーバインドヘルプを表示する
4. When ユーザーが`q`キーでアプリケーションを終了した時, the gh-tuissue shall ターミナルの画面をクリアし、TUIの表示が残らないようにする
5. The gh-tuissue shall 各画面の下部にコンテキストに応じたキーバインドヒントを表示する

### Requirement 9: データ同期・リフレッシュ

**Objective:** As a 開発者, I want Issue一覧を最新の状態に更新したい, so that 他のメンバーの変更を反映した状態で作業できる

#### Acceptance Criteria

1. When ユーザーがリフレッシュ操作を実行した時, the gh-tuissue shall GitHub APIから最新のIssueデータを取得して表示を更新する
2. While データ取得中の間, the gh-tuissue shall ローディングインジケータを表示する
3. If GitHub APIへのリクエストがタイムアウトした場合, the gh-tuissue shall エラーメッセージを表示し、キャッシュされたデータでの表示を維持する

### Requirement 10: プロジェクト紐付けと永続化

**Objective:** As a 開発者, I want リポジトリに対してProjectを一度紐付けたら次回以降自動で使用したい, so that 毎回プロジェクト番号を指定する手間を省ける

#### Acceptance Criteria

1. When リポジトリで初めてgh-tuissueを起動した時（設定ファイルが存在しない場合）, the gh-tuissue shall まず「Projectを紐付けますか？」のYes/No選択UIを表示する
2. When ユーザーがYesを選択した時, the gh-tuissue shall リポジトリに紐づくGitHub Projects V2の一覧を取得し、紐付けるProjectを選択するUIを表示する
3. When ユーザーがNoを選択した時, the gh-tuissue shall Projectを紐付けずにOpen/Closedの2カラムのみで表示する
4. When ユーザーがProjectを選択した時, the gh-tuissue shall 選択結果をリポジトリルートの設定ファイル（`.gh-tuissue.json`）に保存する
5. When 設定ファイルにProject番号が保存されている時, the gh-tuissue shall 次回起動時にProject選択UIを表示せず、保存された設定を自動的に使用する
6. When ユーザーが`--project`フラグを指定した時, the gh-tuissue shall 設定ファイルの値より優先してそのProject番号を使用する（設定ファイルは上書きしない）
7. The gh-tuissue shall 設定変更コマンド（`--config`フラグ等）を提供し、紐付けるProjectを後から変更可能にする
8. When 設定ファイルのProject番号が無効になった場合（Projectが削除された等）, the gh-tuissue shall エラーメッセージを表示し、Project選択UIを再表示する

### Requirement 11: カラム表示・非表示の切り替え

**Objective:** As a 開発者, I want カンバンボードのカラム（ステータス）の表示・非表示を切り替えたい, so that 関心のあるステータスのみに集中して作業できる

#### Acceptance Criteria

1. When ユーザーがカラム非表示操作を実行した時, the gh-tuissue shall 現在フォーカスしているカラムを非表示にし、残りのカラムでレイアウトを再調整する
2. When ユーザーがカラム表示復元操作を実行した時, the gh-tuissue shall 非表示にしたカラムを元の位置に復元する
3. When カラムが非表示の状態で, the gh-tuissue shall 非表示カラム数をステータスバーに表示する
4. The gh-tuissue shall 少なくとも1つのカラムが常に表示された状態を維持する（全カラム非表示を防ぐ）

### Requirement 12: Issue詳細画面からのプロパティインライン編集

**Objective:** As a 開発者, I want Issue詳細画面からラベル・アサイニー・マイルストーンをインラインで編集したい, so that フィルタパネルに遷移せずにプロパティを変更できる

#### Acceptance Criteria

1. When ユーザーがIssue詳細画面でラベル編集操作（lキー）を実行した時, the gh-tuissue shall リポジトリのラベル一覧からの選択UIを直接表示し、変更をGitHubに反映する
2. When ユーザーがIssue詳細画面でアサイニー編集操作（aキー）を実行した時, the gh-tuissue shall コラボレーター一覧からの選択UIを直接表示し、変更をGitHubに反映する
3. When ユーザーがIssue詳細画面でマイルストーン編集操作（mキー）を実行した時, the gh-tuissue shall マイルストーン一覧からの選択UIを直接表示し、変更をGitHubに反映する
4. When 編集が完了または取り消された時, the gh-tuissue shall Issue詳細画面に戻り、更新後のプロパティを反映する
