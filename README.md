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

## Dockerでの実行

`gh release download` でGitHub Releaseのwheelを取得してビルドする

バージョンはDockerfileの `ARG PYTHON_PACKAGE_TAG=v3` で管理します。
CLIの更新時はこの値を変更してイメージを再ビルドします。

BuildKit対応のDockerが必要です。
この構成では公開・非公開どちらのReleaseでも、GitHubトークンをBuildKit secretとして渡します。
Fine-grained PATでは対象リポジトリの Contents: Read 権限を指定します。

別リポジトリのprivate Releaseには、このリポジトリの `GITHUB_TOKEN` だけでは
通常アクセスできないため、PATまたはGitHub Appのトークンを用意してください。

別のReleaseを一時的に試す場合は、ビルド引数で上書きできます。

```sh
GH_TOKEN="$(gh auth token)" docker build \
  --secret id=github_token,env=GH_TOKEN \
  --build-arg PYTHON_PACKAGE_TAG=v2 \
  -t use-python-private-package:v2 .
```

タイムアウトは実行時に変更できます。

```sh
docker run --rm -e PYTHON_CLI_TIMEOUT=1m use-python-private-package -rows 100
```

## GitHub Releaseへのイメージ保存

`v*` のフォーマットでタグ作ってリリースを生やすと、image をビルドし、リリースの Assets にイメージを添付する

workflowはタグが指すコミットのコードを `linux/amd64` 向けにビルドします。
成功するとイメージを `.tar.gz` に保存し、SHA-256チェックサムと一緒に既存のReleaseのAssetsへ添付します。
Releaseのタイトル・本文・Pre-release設定は変更しません。
失敗時はActions画面から再実行できます。同名のAssetsがある場合は置き換えます。

別リポジトリの非公開Releaseを取得する場合は、 `PYTHON_PACKAGE_TOKEN` を登録します。
対象のPythonリポジトリを読めるPATまたはGitHub Appのトークンを指定してください。

ReleaseのAssetsから `use-python-private-package-linux-amd64.tar.gz` をダウンロードして使えます。

```sh
docker load -i use-python-private-package-linux-amd64.tar.gz
docker run --rm --platform linux/amd64 use-python-private-package:v0.1.0-poc -rows 100
```

## 検証

```sh
go test -v ./...
```

CSVの生成、行数の検証に加え、`.venv` の実際のPython CLIで合計計算と
エラーの受け渡しを確認します。CLIがない環境ではPythonとの結合テストをスキップします。
