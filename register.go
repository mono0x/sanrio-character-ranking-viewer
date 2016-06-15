package main

import "log"

type Register struct{}

func (r *Register) Help() string {
	return `sanrio-character-ranking-viewer ranking`
}

func (r *Register) Run(args []string) int {
	context, err := NewAppContext()
	if err != nil {
		log.Fatal(err)
	}
	defer context.Close()

	rankingId := 2

	names := []string{
		`ポムポムプリン`,
		`シンガンクリムゾンズ`,
		`シナモロール`,
		`ぐでたま`,
		`マイメロディ`,
		`プラズマジカ`,
		`ハローキティ`,
		`トライクロニカ`,
		`リトルツインスターズ`,
		`徒然なる操り霧幻庵（つれづれなるあやつりむげんあん）`,
		`KIRIMIちゃん.`,
		`YOSHIKITTY`,
		`クリティクリスタ`,
		`クロミ`,
		`ポチャッコ`,
		`マイ スウィート ピアノ`,
		`けろけろけろっぴ`,
		`タキシードサム`,
		`ウィッシュミー メル`,
		`バッドばつ丸`,
		`ターフィー`,
		`マロンクリーム`,
		`ぼんぼんりぼん`,
		`ジュエルペット`,
		`チャーミーキティ`,
		`コロコロクリリン`,
		`チアリーチャム`,
		`ルロロマニック`,
		`みんなのたあ坊`,
		`ハンギョドン`,
		`ディアダニエル`,
		`パティ＆ジミー`,
		`シュガーバニーズ`,
		`ウサハナ`,
		`ニャ ニィ ニュ ニェ ニョン`,
		`タイニーチャム`,
		`おさるのもんきち`,
		`ハミングミント`,
		`あひるのペックル`,
		`マシュマロみたいなふわふわにゃんこ`,
		`ゴロピカドン`,
		`イチゴマン`,
		`マイマイ`,
		`パタパタペッピー`,
		`シンカイゾク`,
		`シンカンセン`,
		`かしわんこもち`,
		`ぽこぽん日記`,
		`ミスターメン リトルミス`,
		`ハニーモモ`,
		`ぱんくんち`,
		`いちごの王さま`,
		`ボタンノーズ`,
		`笑う女`,
		`てのりくま`,
		`ウシ`,
		`ウメ屋雑貨店`,
		`ザ ボードビルデュオ`,
		`ザシキブタ`,
		`カカオとバニラ`,
		`シュガーメヌエット`,
		`ザ ラナバウツ`,
		`チェリーナチェリーネ`,
		`たらいぐまのらんどりー`,
		`セブンシリードワーフ`,
		`ちょボット`,
		`ドリームテイルクーベア`,
		`ニョッキ＆ペンネ`,
		`ぽんぽんひえた`,
		`センゴクプリズン`,
		`ダークグレープマン`,
		`プワワ`,
		`ウィンキーピンキー`,
		`フランボアルゥルゥ`,
		`ちびまる`,
		`スウィートコロン`,
		`バニー＆マッティ`,
		`ひきだしあいた`,
		`チョコキャット`,
		`ミミックマイク`,
		`リスル`,
		`ピーター デイビス`,
		`スポーティングベアーズ`,
		`リルリルフェアリル`,
		`歯ぐるまんすたいる`,
		`中年ひろいんＯｊｉｓａｎ’ｓ`,
		`お酒たちの日常`,
		`アグレッシブ烈子`,
		`ペペペペン議員`,
		`ちんじゅうみん＆ゴーちゃん。`,
		`ふくちゃん`,
		`リトルラヴィン`,
		`タイニーポエム`,
		`だちょのすけ`,
		`ノラネコランド`,
		`パウピポ`,
		`フレッシュパンチ`,
		`八千代チャーマー`,
		`ララバイラバブルズ`,
		`るるる学園`,
	}

	for _, name := range names {
		character, err := FindCharacterByName(context.dbMap, name)
		if err != nil {
			character = &Character{
				Name: name,
			}
			if err := context.dbMap.Insert(character); err != nil {
				log.Fatal(err)
			}
		}
		var entry Entry
		if err := context.dbMap.SelectOne(&entry, `
			SELECT * FROM entry
			WHERE
				ranking_id = :ranking_id
				AND character_id := character_id
			LIMIT 1
		`, map[string]string{
			"ranking_id":   string(rankingId),
			"character_id": string(character.Id),
		}); err != nil {
			entry = Entry{
				RankingId:   rankingId,
				CharacterId: character.Id,
			}
			if err := context.dbMap.Insert(&entry); err != nil {
				log.Fatal(err)
			}
		}
	}

	return 0
}

func (r *Register) Synopsis() string {
	return `Register characters to the ranking`
}
