# amachobo
- Amazonで事業主借で決済したものをfreeeに簡単に登録するやつ

## Usage
- Install Chrome Extension [アマゾン注文履歴フィルタ](https://chrome.google.com/webstore/detail/%E3%82%A2%E3%83%9E%E3%82%BE%E3%83%B3%E6%B3%A8%E6%96%87%E5%B1%A5%E6%AD%B4%E3%83%95%E3%82%A3%E3%83%AB%E3%82%BF/jaikhcpoplnhinlglnkmihfdlbamhgig/related?hl=ja&gl=JP)
- Export All CSVs from Amazon(Currently, only amazon.co.jp is supported)
- Execute amachobo with the path to CSVs
```console
$ amachobo ./path-to-digital.csv ./path-to-non-digital.csv ...
```
- Generated CSV for Freee will be saved in freee-*.csv
