# GoからPython CLIを呼び出すPoC

Goで `ID,price,quantity` のCSVを生成し、インストール済みのPython CLIを
`os/exec` で子プロセスとして起動します。Pythonが出力ファイルに書いた
「単価 × 個数」の合計をGoが読み取り、標準出力します。

## 環境の準備

Go 1.25以上、uv、隣の `../python-private-package` が必要です。
すでに `.venv` にCLIをインストールしている場合、この操作は不要です。

```sh
uv venv
uv pip install --python .venv/bin/python ../python-private-package
```

`.python-version` でPython 3.13を指定しています。
`.venv` は各環境で作成し、Gitには含めません。

## 実行

このリポジトリのルートで実行します。venvのactivateや `uv run` は不要です。

```sh
go run . -rows 100
```

`-rows` はヘッダーを除くデータ行数で、省略時は10行です。
IDは `A001` からの連番、単価は1.00〜1000.00、個数は1〜10でランダムに生成します。
標準出力には合計金額だけを出力するため、実行ごとに結果は変わります。
CSVと計算結果は一時ディレクトリで受け渡し、終了時に削除します。
CLIの実行はデフォルトで30秒でタイムアウトします。エラーは標準エラー出力に表示し、終了コード1で終了します。

## 環境ごとの設定

`config.go` で環境変数を読み取ります。起動ディレクトリに `.env` があれば、
`godotenv` で読み込みます。ローカル開発ではサンプルをコピーして編集できます。

```sh
cp .env.example .env
```

```dotenv
PYTHON_CLI_PATH=.venv/bin/python-private-package
PYTHON_CLI_TIMEOUT=30s
```

| 環境変数 | 内容 | 未設定・空文字の場合 |
| --- | --- | --- |
| `PYTHON_CLI_PATH` | Python CLIの実行ファイル | `.venv/bin/python-private-package` |
| `PYTHON_CLI_TIMEOUT` | CLI実行の制限時間。`30s`、`1m` などの正の時間 | `30s` |

優先順位は「既存の環境変数 → `.env` → デフォルト値」です。
既存の環境変数が空文字なら `.env` の値は使わず、デフォルト値を使います。
`.env` がなくても起動できます。不正な `.env` やタイムアウト値は起動時にエラーになります。
`.env` はGit管理から除外し、共有する設定例は `.env.example` で管理します。

本番では起動環境から設定を渡します。CLIのパスは絶対パスにすると、起動ディレクトリに依存しません。

```sh
PYTHON_CLI_PATH=/opt/app/.venv/bin/python-private-package \
PYTHON_CLI_TIMEOUT=1m \
go run . -rows 100
```

`-cli` フラグは使いません。行数は引き続き `-rows` で指定します。

## 検証

```sh
go test -v ./...
```

CSVの生成、行数の検証に加え、`.venv` の実際のPython CLIで合計計算と
エラーの受け渡しを確認します。CLIがない環境ではPythonとの結合テストをスキップします。
