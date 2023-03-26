// TODO: parse a
// give locale files zh-CN.ts, find the mssing key in zh-Tw.ts, en-US.ts. the local files are export default dictionary of dictionary.
import zhCn from './zh-CN'
import zhTw from './zh-TW'
import enUs from './en-US'

// rome-ignore lint/suspicious/noExplicitAny: <explanation>
const findMissingKeys = (base: Record<string, any>, other: Record<string, any>) => {
  // rome-ignore lint/suspicious/noExplicitAny: <explanation>
  const missingKeys: Record<string, any> = {}

  for (const key in base) {
    // eslint-disable-next-line no-prototype-builtins
    if (base.hasOwnProperty(key)) {
      if (!other[key]) {
        missingKeys[key] = base[key]
      }
      else if (typeof base[key] === 'object' && typeof other[key] === 'object') {
        const subMissingKeys = findMissingKeys(base[key], other[key])
        if (Object.keys(subMissingKeys).length > 0)
          missingKeys[key] = subMissingKeys
      }
    }
  }

  return missingKeys
}

// Find missing keys in zh-TW
const zhTwMissingKeys = findMissingKeys(zhCn, zhTw)
console.log('\n\n please translate to zh-TW:', zhTwMissingKeys)

// Find missing keys in en-US
const enUsMissingKeys = findMissingKeys(zhCn, enUs)
console.log('\n\n please translate to en-US:', enUsMissingKeys)
