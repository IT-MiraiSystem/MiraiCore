# MiraiCore
## 説明書きは後回し

## MySQLについて
コンテナ内でDBに接続するのと外部から接続するのではわけが違うらしい
コンテナ内ホストネーム:db
コンテナ外ホストネーム:コンテナのIPアドレス

その他外部からのアクセスに必要なオプション
```
--protocol tcp
```
## サーバーログについて
log/MiraiCore-API.logにサーバーログを記載する。
500MBまで貯まると自動的にもう一つファイルが作られる
## ビルド方法
```sh
docker compose up -d
```
## ビルドした際のURLなど
[API Base URL](http://localhost/api)
[API Document](http://localhost/document/api)
[教師用アプリ](http://localhost/teacher)
### tailコマンドはいいぞ
``` sh
sudo tail -f log/MiraiCore-API.log
```
これを使うと自動的にログファイルを更新されていちいちコマンドを打たなくて住むから
皆これ使おうぜ