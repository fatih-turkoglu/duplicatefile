// Dil seçimi (saf mantık; node --test ile sınanır).

export const AUTO = 'auto'

// selectLanguage, kullanılacak dil dosyasının kodunu döner; 'en' gömülü
// İngilizcedir (en.lng olmasa da kullanılabilir).
//  - preference bir dil koduysa ve dosyası varsa o.
//  - 'auto' (ya da dosyası artık olmayan bir tercih) ise Windows'un dil
//    sırasında ilk eşleşen: önce tam kod (pt-br), sonra ana dil (pt), sonra
//    aynı ana dilin başka bir varyantı (pt-PT sistem → pt-br dosyası).
//    Sırada İngilizce dosyası olan dillerden önce gelirse İngilizce seçilir.
//  - Hiçbiri yoksa İngilizce.
export function selectLanguage(preference: string, system: string[], available: string[]): string {
  const avail = available.map((a) => a.toLowerCase())
  const pref = preference.toLowerCase()
  if (pref !== AUTO && (avail.includes(pref) || pref === 'en')) return pref

  for (const s of system) {
    const tag = s.toLowerCase()
    const primary = tag.split('-')[0]
    if (avail.includes(tag)) return tag
    if (avail.includes(primary)) return primary
    const variant = avail.find((a) => a.split('-')[0] === primary)
    if (variant) return variant
    if (primary === 'en') return 'en'
  }
  return 'en'
}
