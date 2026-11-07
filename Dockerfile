# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go config.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/app .

FROM python:3.13-slim-bookworm AS python-install

# 利用するReleaseのバージョンはここで管理する。
ARG PYTHON_PACKAGE_REPOSITORY=t0ratiger/python-private-package
ARG PYTHON_PACKAGE_TAG=v3

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates gh \
    && rm -rf /var/lib/apt/lists/*

# 同じGitHubトークンを、このRUNの間だけGH_TOKENとして参照する。
RUN --mount=type=secret,id=github_token,env=GH_TOKEN,required=true \
    gh release download "${PYTHON_PACKAGE_TAG}" \
      --repo "${PYTHON_PACKAGE_REPOSITORY}" \
      --pattern 'python_private_package-*-py3-none-any.whl' \
      --dir /tmp/wheels

RUN python -m venv /opt/venv \
    && /opt/venv/bin/python -m pip install --disable-pip-version-check --no-cache-dir /tmp/wheels/*.whl \
    && rm -rf /tmp/wheels

FROM python:3.13-slim-bookworm AS runtime
WORKDIR /app
COPY --from=go-build /out/app /app/app
COPY --from=python-install /opt/venv /opt/venv
ENV PYTHON_CLI_PATH=/opt/venv/bin/python-private-package \
    PYTHON_CLI_TIMEOUT=30s
USER 10001:10001
ENTRYPOINT ["/app/app"]
