// lang/en.lng dosyasını gömülü İngilizce metinlerden (src/i18n/en.ts) üretir.
// en.lng çevirmenler için şablondur; uygulama İngilizceyi zaten içinde taşır.
// Kullanım (frontend klasöründe): npm run lang:en
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const { en } = await import(pathToFileURL(path.join(here, '..', 'src', 'i18n', 'en.ts')).href)
const out = path.join(here, '..', '..', 'lang', 'en.lng')

const header = `# DuplicateFile language file — English (template for translations)
#
# This file is generated from frontend/src/i18n/en.ts (npm run lang:en);
# English is built into the program, so this file is only a template.
#
# How to translate:
#   1. Copy this file and name it after the language code, e.g. it.lng, es.lng, pt-br.lng
#   2. Set @name (the language's own name) and @locale (for number/date formats)
#   3. Translate the text after "=". Keep the keys and {placeholders} unchanged.
#   4. Put the file into the "lang" folder next to DuplicateFile.exe.
#
# Format: UTF-8, one "key = text" per line. Lines starting with # are comments.
#   \\n = new line.  Missing keys are shown in English.
#   Number-dependent texts have plural suffixes; use the forms your language needs:
#   .zero .one .two .few .many .other  (.other is required). See CLDR plural rules.

@name = English
@locale = en-US
`

const escape = (s) => s.replaceAll('\\', '\\\\').replaceAll('\n', '\\n')
let body = ''
let prevGroup = ''
for (const [key, value] of Object.entries(en)) {
  const group = key.split('.')[0]
  if (group !== prevGroup) {
    body += '\n'
    prevGroup = group
  }
  body += `${key} = ${escape(value)}\n`
}

fs.mkdirSync(path.dirname(out), { recursive: true })
fs.writeFileSync(out, header + body, 'utf8')
console.log(`${out}: ${Object.keys(en).length} anahtar`)
