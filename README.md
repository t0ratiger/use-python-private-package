# GoからPython CLIを呼び出すPoC

GoでCSVを生成し、インストール済みのPython CLIを `os/exec` で子プロセスとして起動。
Pythonが出力ファイルに書いた「単価 × 個数」の合計をGoが読み取り、標準出力する。

## 環境の準備

- Go
- uv
- [python-private-package](https://github.com/t0ratiger/python-private-package)

python-private-package をクローンして、インストールする

```sh
uv venv
uv pip install --python .venv/bin/python path/to/python-private-package
```

パッケージのインストールは、`-e` フラグをつけてリポジトリのルートを指定すると、editable インストールができ、コードの変更を再インストールなしで反映可能。

whl ファイルを指定すれば、ビルド済みのパッケージをインストールできる。

## 実行

```sh
go run . -rows 100
```

- venvのactivateや `uv run` は不要。
- `-rows` はヘッダーを除くデータ行数で、省略時は10

IDは `A001` からの連番、単価は1.00〜1000.00、個数は1〜10でランダムに生成します。
標準出力は実行ごとに結果は変わります。
CSVと計算結果は一時ディレクトリで受け渡し、終了時に削除します。

## 環境ごとの設定

cli のパスとタイムアウトの秒数の設定が可能。cli のパスは、venv/bin の中にある

```sh
cp .env.example .env
```

```dotenv
PYTHON_CLI_PATH=.venv/bin/python-private-package
PYTHON_CLI_TIMEOUT=30s
```

本番では起動環境から設定を渡すようにする

```sh
PYTHON_CLI_PATH=/opt/app/.venv/bin/python-private-package \
PYTHON_CLI_TIMEOUT=1m \
go run . -rows 100
```

## 検証

```sh
go test -v ./...
```

CSVの生成、行数の検証に加え、`.venv` の実際のPython CLIで合計計算と
エラーの受け渡しを確認します。CLIがない環境ではPythonとの結合テストをスキップします。
