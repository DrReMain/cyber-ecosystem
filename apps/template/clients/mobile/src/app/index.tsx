import { Text, View } from "react-native"

export default function Index() {
  return (
    <View className="flex-1 items-center justify-center bg-gray-50 dark:bg-gray-950">
      <Text className="font-bold text-2xl text-gray-950 dark:text-gray-50">
        Hello World
      </Text>
      <Text className="text-gray-700 text-sm dark:text-gray-300">
        React Native
      </Text>
    </View>
  )
}
