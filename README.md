[![CircleCI](https://circleci.com/gh/RoboCup-SSL/ssl-vision-client/tree/master.svg?style=svg)](https://circleci.com/gh/RoboCup-SSL/ssl-vision-client/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/RoboCup-SSL/ssl-vision-client?style=flat-square)](https://goreportcard.com/report/github.com/RoboCup-SSL/ssl-vision-client)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/RoboCup-SSL/ssl-vision-client/pkg/vision)
[![Release](https://img.shields.io/github/release/RoboCup-SSL/ssl-vision-client.svg?style=flat-square)](https://github.com/RoboCup-SSL/ssl-vision-client/releases/latest)

# ssl-vision-client

A graphical client for [ssl-vision](https://github.com/RoboCup-SSL/ssl-vision) that receives multicast packages and
shows them in a web-ui.

## このフォークについて

[RoboCup-SSL/ssl-vision-client](https://github.com/RoboCup-SSL/ssl-vision-client) を ibis-ssl チームがフォークしたものです。元リポジトリの成果に敬意を表しつつ、以下の機能を独自に追加しています。

### 追加機能

- **ログリプレイ** — ログファイルのドラッグ&ドロップ読み込み、シークバー操作、キーボードショートカット対応
- **grSim 連携** — フィールド上でボール・ロボットをダブルクリックで移動操作
- **Auto Ball Placement 自動化** — ボールプレースメントの自動実行
- **ゲームイベント対応** — AutoRef / GC イベントのフル対応とトースト通知（リプレイ中も表示）
- **UI 刷新** — Material 3 Expressive デザイン、設定パネルをサイドドロワーに統合
- **接続管理改善** — データソース切替、パケット受信の自動検出・自動切替

## Usage
If you just want to use this app, simply download the latest [release binary](https://github.com/RoboCup-SSL/ssl-vision-client/releases/latest).
The binary is self-contained. No dependencies are required.

You can also use pre-build docker images:
```shell script
docker pull robocupssl/ssl-vision-client
docker run -p 8082:8082 robocupssl/ssl-vision-client
```

By default, the UI is available at http://localhost:8082

## Development

### Requirements

You need to install following dependencies first:

* Go
* Node

See [.circleci/config.yml](.circleci/config.yml) for compatible versions.

### Frontend

See [frontend/README.md](frontend/README.md)

### Build

Build and install all binaries:

```bash
make install
```

### Run

Run the backend:

```bash
go run cmd/ssl-vision-client/main.go
```

### Update generated protobuf code

Generate the code for the `.proto` files after you've changed anything in a `.proto` file with:

```shell
make proto
```
