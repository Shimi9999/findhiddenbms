# findhiddenbms

指定したフォルダ内の隠しBMS譜面を探します。

検索対象:
- BMS譜面を内包するテキストファイル(ノーツ数0は除く)
- 隠し譜面の存在を示唆するテキストファイル
- 隠し譜面の可能性のあるファイル名のファイル
- Zipファイル(無配置BMSを示唆するファイル名は除く)

## Usage
```
findhiddenbms <dirpath>
```

## Exmaple
```
> findhiddenbms exampleFolder
*hasBMS: exampleFolder\Song1\chart_kakushi.txt
**TargetFilename: exampleFolder\Song2\secret_fumen.bm_
*Zipfile: exampleFolder\Song3\for_expert.zip
**ContainsHiddenWords: 隠し exampleFolder\Song4\_readme.txt
```