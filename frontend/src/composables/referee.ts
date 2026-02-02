import { type Referee, RefereeSchema } from '@/proto/gc/ssl_gc_referee_message_pb.ts'
import { useWebSocketProtobuf } from '@/composables/websocket.ts'
import { computed } from 'vue'

export const useReferee = () => {
  const { message: referee, status } = useWebSocketProtobuf<Referee>('/api/referee', RefereeSchema)
  const connected = computed(() => status.value === 'OPEN')
  return { referee, connected }
}
