import { useAtom } from "jotai"
import { Modal, Pressable, Text, View } from "react-native"
import { LOCALES, type Locale, SUPPORTED_LOCALES } from "@/lib/i18n"
import { localeAtom, setLocaleAtom } from "@/stores/i18n/store"

interface LocalePickerProps {
  visible: boolean
  onClose: () => void
}

export function LocalePicker({ visible, onClose }: LocalePickerProps) {
  const [locale] = useAtom(localeAtom)
  const [, setLocale] = useAtom(setLocaleAtom)

  const handleSelect = (l: Locale) => {
    setLocale(l)
    onClose()
  }

  return (
    <Modal
      animationType="fade"
      onRequestClose={onClose}
      transparent
      visible={visible}
    >
      <Pressable className="flex-1 bg-black/40" onPress={onClose}>
        <View className="absolute right-0 bottom-0 left-0 rounded-t-2xl bg-white pt-3 pb-8 dark:bg-gray-900">
          <View className="mb-3 h-1 w-10 self-center rounded-full bg-gray-300 dark:bg-gray-600" />
          {SUPPORTED_LOCALES.map((l) => (
            <Pressable
              className="px-6 py-3 active:bg-gray-100 dark:active:bg-gray-800"
              key={l}
              onPress={() => handleSelect(l)}
            >
              <Text
                className={`text-base ${
                  l === locale
                    ? "font-semibold text-indigo-500"
                    : "text-gray-700 dark:text-gray-300"
                }`}
              >
                {LOCALES[l].label}
              </Text>
            </Pressable>
          ))}
        </View>
      </Pressable>
    </Modal>
  )
}
