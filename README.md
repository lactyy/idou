# idou

A server tool that allows to you join any external server from friend list.

I heard a rumor that a world allows to connect external servers are no longer usable since 1.21.20, so I made one.
1.21.20 でコンソールなどで特集サーバー以外に接続するワールドが使えなくなった, という噂を聞いたので作ってみました. とても適当です.

This server/tool can be hosted just like opening a world, with no port forwarding required to other players to join.
このサーバー/ツールは手軽にワールドを開く感覚で, ポート開放不要で使用できます.

Compared to other methods (like DNS forwarding), You can just add a friend, and that's it.
従来の方法 (DNSなど) に比べて, フレンドを追加するだけで大丈夫です.

## 使い方 / How to use
If you have Go installed on your device, you can run ``go install github.com/lactyy/idou@latest``, then run ``idou`` from command-line.

If you haven't installed Go on your device, you can download an executable from Releases tab, then run ``./idou``.

The location where you ran the command will be the data folder, which contains a database for remembering servers which players has added, and an auth.cred file that stores your credentials.

## 警告 / Warning
Your IP addresses will be broadcasted to the players connecting to the world via signaling. It is highly recommended to host this server on a separated network.

Please use a separate account for hosting this, to avoid getting flagged.

## Credits
Thanks to the authors of [df-mc/nethernet-spec](https://github.com/df-mc/nethernet-spec) to writing the specification.

A huge thanks to Da1z981 for joining the debugging, and the community for helping.