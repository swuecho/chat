python ./scripts/merge_keys.py web/src/locales/en-US.json web/src/locales/en-US-more.json > ./web/src/locales/en-USx.json
mv ./web/src/locales/en-USx.json ./web/src/locales/en-US.json
python ./scripts/merge_keys.py web/src/locales/zh-TW.json web/src/locales/zh-TW-more.json > ./web/src/locales/zh-TWx.json
mv ./web/src/locales/zh-TWx.json ./web/src/locales/zh-TW.json