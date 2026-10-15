import { useEffect, useState } from 'react'

// localStorage'da saklanan nesne durumu. Kayıtlı veri varsayılanla
// birleştirilir; böylece sonradan eklenen alanlar da varsayılan değerini alır.
// Okuma/yazma hatalarında (bozuk veri, kota) varsayılanla sessizce devam eder.
export function useLocalStorage<T extends object>(key: string, initial: T): [T, React.Dispatch<React.SetStateAction<T>>] {
  const [value, setValue] = useState<T>(() => {
    try {
      const raw = localStorage.getItem(key)
      return raw ? { ...initial, ...JSON.parse(raw) } : initial
    } catch {
      return initial
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(key, JSON.stringify(value))
    } catch {
      // saklanamazsa yalnızca bu oturumda geçerli olur
    }
  }, [key, value])

  return [value, setValue]
}
