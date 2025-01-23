# MiraiCore
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FIT-MiraiSystem%2FMiraiCore.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2FIT-MiraiSystem%2FMiraiCore?ref=badge_shield&issueType=license)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FIT-MiraiSystem%2FMiraiCore.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2FIT-MiraiSystem%2FMiraiCore?ref=badge_shield&issueType=security)
## MiraiGateのためのAPIサーバー
MiraiGateを作るためのAPIサーバー
## 実行に必要なファイル
Google Drive上のDB/IT未来在学生.xlsx → db/IT未来在学生.xlsx \
Firebaseの認証情報(秘密鍵) 秘密鍵のJSON → config/FirebaseConfig.json

### オプション的なファイル
授業情報を登録したい！ DB/授業登録.sql → db/授業登録

## MySQLについて
コンテナ内でDBに接続するのと外部から接続するのではわけが違うらしい \
コンテナ内ホストネーム:db \
コンテナ外ホストネーム:コンテナのIPアドレス \

その他外部からのアクセスに必要なオプション
```
--protocol tcp
```
## サーバーログについて
log/MiraiCore-API.logにサーバーログを記載する。 \
500MBまで貯まると自動的にもう一つファイルが作られる
## ビルド方法
```sh
docker compose up -d
```
## ビルドした際のURLなど
[API Base URL](http://localhost/api) \
[API Document](http://localhost/document/api) \
[教師用アプリ](http://localhost/teacher) \
### tailコマンドはいいぞ
``` sh
sudo tail -f log/MiraiCore-API.log
```
これを使うと自動的にログファイルを更新されていちいちコマンドを打たなくて済むから
皆これ使おうぜ
