package main

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// TODO: Implement reading catalog from JSON files

func init() {
	for tag, dictionary := range map[language.Tag]map[string]string{
		language.English: {
			"menu.title": "Menu",

			"item.menu": "Open Menu",

			"menu.add.title":         "Add a server",
			"menu.add.button":        "Add a server",
			"menu.add.alreadyExists": "§cThis server has already been added.",
			"menu.add.success":       "§aThe server has been added.",

			"menu.remove.title":   "Remove a server",
			"menu.remove.button":  "Remove a server",
			"menu.remove.success": "§cThe server has been removed.",

			"menu.connect.title":         "Connect to a server",
			"menu.connect.button":        "Connect to a server",
			"menu.connect.remember.text": "Remember this server",

			"menu.settings.title":  "Settings",
			"menu.settings.button": "Settings",

			"server.button.text":         "%s",
			"server.button.text.named":   "%s (%s)",
			"server.address.text":        "Server Address",
			"server.address.placeholder": "Address or address:port",
			"server.name.text":           "Server Name (Optional)",

			"server.error.resolve": "§cError resolving server address: %s",
			"server.error.limit":   "§cYou can add up to %d servers.",

			"command.language.error.parse": "§cError parsing language: %s",

			"language.unsupported": "§cUnsupported language: %s",
			"language.changed":     "§aThe language has been changed to %s.",
			"language.auto":        "§aWe have set your language to %s, based on your game language. You can change your language by opening settings.",

			"menu.language.title":           "Change language",
			"menu.language.button":          "Change language",
			"menu.language.dropdown.text":   "Please select a language:",
			"menu.language.dropdown.option": "%s (%s)",
		},
		language.Japanese: {
			"menu.title": "メニュー",

			"item.menu": "メニューを開く",

			"menu.add.title":         "サーバーを追加",
			"menu.add.button":        "サーバーを追加",
			"menu.add.alreadyExists": "§cこのサーバーは既に追加されています.",
			"menu.add.success":       "§aサーバーが追加されました.",

			"menu.remove.title":   "サーバーを削除",
			"menu.remove.button":  "サーバーを削除",
			"menu.remove.success": "§cサーバーが削除されました.",

			"menu.connect.title":         "サーバーに接続",
			"menu.connect.button":        "サーバーに接続",
			"menu.connect.remember.text": "このサーバーを記憶する",

			"menu.settings.title":  "設定",
			"menu.settings.button": "設定",

			"server.button.text":         "%s",
			"server.button.text.named":   "%s (%s)",
			"server.address.text":        "サーバーのアドレス",
			"server.address.placeholder": "アドレス, または アドレス:ポート",
			"server.name.text":           "サーバーの名前 (省略可)",

			"server.error.resolve": "§cサーバーアドレスを解決できませんでした: %s",
			"server.error.limit":   "§c追加できるサーバーは最大 %d 個までです.",

			"command.language.error.parse": "§c無効な言語です: %s",

			"language.unsupported": "§cサポートされていない言語です: %s",
			"language.changed":     "§a言語を %s に変更しました.",
			"language.auto":        "§aゲーム内の言語に基づいて, 言語を %s に変更しました. メニューから設定を開くことで, 言語を変更できます.",

			"menu.language.title":           "言語を変更",
			"menu.language.button":          "言語を変更",
			"menu.language.dropdown.text":   "言語を選択してください:",
			"menu.language.dropdown.option": "%s (%s)",
		},
	} {
		for key, msg := range dictionary {
			if err := message.SetString(tag, key, msg); err != nil {
				panic(fmt.Sprintf("%s: error registering catalog message: %q (%q)", tag, key, msg))
			}
		}
	}
}
