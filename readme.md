# ChatRoom

## 說明

範例通訊軟體 API，目前已實現以下功能:

+ 登入/註冊/更改用戶訊息

+ 以 room 為單位進行通話

+ room 有 admin 可以管理成員進出，user 也可以申請加入 room

+ 支援貼圖功能 sticker : message 字串如果有包含 sticker::id1::id2，
  會交由後端認證，會去除不存在或是尚未購買的sticker後，儲存在資料庫。之後前端fetch
  message 後會根據  sticker::id1::id2 顯示貼圖。

+ 有 wallet 模擬充值功能，暫時只能用來買 sticker 

## 用到的技術

gin, gorm, postgresql, redis

## TODO 

+ 將每個 user 的 sticker 資訊存到 redis
