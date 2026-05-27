import { App } from "antd"
import { useEffect } from "react"
import { onApiError } from "./error-events"

export function AntdErrorFeedbackAdapter() {
  const { message, notification } = App.useApp()

  useEffect(() => {
    return onApiError((error, type) => {
      if (type === "mutation") {
        void message.error(error.message)
      } else {
        notification.error({ title: error.message })
      }
    })
  }, [message, notification])

  return null
}
