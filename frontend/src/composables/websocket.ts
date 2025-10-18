import { fromBinary, type Message } from '@bufbuild/protobuf'
import type { GenMessage } from '@bufbuild/protobuf/codegenv1'
import { computed, type MaybeRefOrGetter, toValue } from 'vue'
import { computedAsync, useWebSocket } from '@vueuse/core'
import { determineWebSocketAddress } from '@/helpers/websocket.ts'

export const useWebSocketProtobuf = <T extends Message>(
  path: MaybeRefOrGetter<string>,
  schema: GenMessage<T>,
  enabled?: MaybeRefOrGetter<boolean>,
) => {
  const url = computed(() => {
    if (enabled === undefined || toValue(enabled)) {
      return determineWebSocketAddress(toValue(path))
    } else {
      return undefined
    }
  })
  const { data, status } = useWebSocket<Blob>(url, {
    autoReconnect: true,
    immediate: true,
  })

  const message = computedAsync(
    async () => {
      if (data.value && url.value) {
        const blob = await data.value.arrayBuffer()
        const bytes = new Uint8Array(blob)
        const msg = fromBinary(schema, bytes)
        return msg
      }
      return undefined
    },
    undefined,
    { lazy: false, evaluating: undefined }
  )

  return { message, status }
}
