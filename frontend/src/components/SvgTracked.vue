<script setup lang="ts">
import SvgRobot from '@/components/SvgRobot.vue'
import SvgBall from '@/components/SvgBall.vue'
import SvgText from '@/components/SvgText.vue'
import type { TrackedFrame } from '@/proto/tracked/ssl_vision_detection_tracked_pb.ts'
import type { RobotId } from '@/proto/gc/ssl_gc_common_pb.ts'
import { Team } from '@/proto/gc/ssl_gc_common_pb.ts'
import { computed } from 'vue'

const props = defineProps<{
  trackedFrame: TrackedFrame
  showVelocity: boolean
  showKickedBall: boolean
  showVisibility: boolean
}>()

const VEL_SCALE = 0.15
const MAX_ARROW_LENGTH = 0.5
const MIN_BALL_SPEED = 0.1
const MIN_ROBOT_SPEED = 0.05

function speed2d(vx: number, vy: number): number {
  return Math.sqrt(vx * vx + vy * vy)
}

function clampedArrow(baseX: number, baseY: number, vx: number, vy: number) {
  const spd = speed2d(vx, vy)
  const len = Math.min(spd * VEL_SCALE, MAX_ARROW_LENGTH)
  const scale = len / spd
  return { endX: baseX + vx * scale, endY: -baseY - vy * scale, speedText: spd.toFixed(1) + 'm/s' }
}

function isKickingRobot(robotId: RobotId): boolean {
  const kicked = props.trackedFrame.kickedBall
  if (!kicked?.robotId) return false
  return kicked.robotId.id === robotId.id && kicked.robotId.team === robotId.team
}

const ballArrows = computed(() =>
  props.trackedFrame.balls.flatMap((b) => {
    if (!b.vel) return []
    const spd = speed2d(b.vel.x, b.vel.y)
    if (spd <= MIN_BALL_SPEED) return []
    const arrow = clampedArrow(b.pos!.x, b.pos!.y, b.vel.x, b.vel.y)
    return [{ x1: b.pos!.x, y1: b.pos!.y, ...arrow }]
  }),
)

const robotArrows = computed(() =>
  props.trackedFrame.robots.flatMap((r) => {
    if (!r.vel) return []
    const spd = speed2d(r.vel.x, r.vel.y)
    if (spd <= MIN_ROBOT_SPEED) return []
    const arrow = clampedArrow(r.pos!.x, r.pos!.y, r.vel.x, r.vel.y)
    const isYellow = r.robotId!.team === Team.YELLOW
    return [{
      x1: r.pos!.x,
      y1: r.pos!.y,
      ...arrow,
      color: isYellow ? '#cccc00' : '#6688ff',
      marker: isYellow ? 'url(#arrowhead-yellow)' : 'url(#arrowhead-blue)',
    }]
  }),
)
</script>

<template>
  <defs>
    <marker
      id="arrowhead-ball"
      markerUnits="userSpaceOnUse"
      markerWidth="0.06"
      markerHeight="0.04"
      refX="0.055"
      refY="0.02"
      orient="auto"
    >
      <polygon points="0 0, 0.06 0.02, 0 0.04" fill="orange" fill-opacity="0.85" />
    </marker>
    <marker
      id="arrowhead-yellow"
      markerUnits="userSpaceOnUse"
      markerWidth="0.06"
      markerHeight="0.04"
      refX="0.055"
      refY="0.02"
      orient="auto"
    >
      <polygon points="0 0, 0.06 0.02, 0 0.04" fill="#cccc00" fill-opacity="0.85" />
    </marker>
    <marker
      id="arrowhead-blue"
      markerUnits="userSpaceOnUse"
      markerWidth="0.06"
      markerHeight="0.04"
      refX="0.055"
      refY="0.02"
      orient="auto"
    >
      <polygon points="0 0, 0.06 0.02, 0 0.04" fill="#6688ff" fill-opacity="0.85" />
    </marker>
  </defs>

  <!-- キックボール予測（背面） -->
  <g v-if="showKickedBall && trackedFrame.kickedBall">
    <line
      v-if="trackedFrame.kickedBall.stopPos"
      :x1="trackedFrame.kickedBall.pos!.x"
      :y1="-trackedFrame.kickedBall.pos!.y"
      :x2="trackedFrame.kickedBall.stopPos.x"
      :y2="-trackedFrame.kickedBall.stopPos.y"
      stroke="red"
      :stroke-width="0.005"
      stroke-dasharray="0.04 0.02"
      stroke-opacity="0.7"
    />
    <circle
      v-if="trackedFrame.kickedBall.stopPos"
      :cx="trackedFrame.kickedBall.stopPos.x"
      :cy="-trackedFrame.kickedBall.stopPos.y"
      r="0.03"
      stroke="red"
      :stroke-width="0.005"
      fill="rgba(255,0,0,0.2)"
    />
  </g>

  <!-- ボール -->
  <SvgBall
    v-for="(s, i) in trackedFrame.balls"
    :key="'tracked-ball-' + i"
    :x="s.pos!.x"
    :y="s.pos!.y"
    :height="(s.pos!.z || 0) * 1000"
    :visibility="showVisibility ? (s.visibility || 1) : 1"
  />

  <!-- ボール速度矢印 -->
  <g v-if="showVelocity">
    <g v-for="(arrow, i) in ballArrows" :key="'ball-vel-' + i">
      <line
        :x1="arrow.x1"
        :y1="-arrow.y1"
        :x2="arrow.endX"
        :y2="arrow.endY"
        stroke="orange"
        :stroke-width="0.008"
        stroke-opacity="0.85"
        marker-end="url(#arrowhead-ball)"
      />
      <SvgText :x="arrow.endX" :y="-arrow.endY" :text="arrow.speedText" color="orange" />
    </g>
  </g>

  <!-- ロボット -->
  <SvgRobot
    v-for="(s, i) in trackedFrame.robots"
    :key="'tracked-robot-' + i"
    :x="s.pos!.x"
    :y="s.pos!.y"
    :orientation="s.orientation"
    :id="s.robotId!.id"
    :team-color="s.robotId!.team === Team.YELLOW ? 'YELLOW' : 'BLUE'"
    :visibility="showVisibility ? (s.visibility || 1) : 1"
    :is-kicker="showKickedBall && isKickingRobot(s.robotId!)"
  />

  <!-- ロボット速度矢印 -->
  <g v-if="showVelocity">
    <g v-for="(arrow, i) in robotArrows" :key="'robot-vel-' + i">
      <line
        :x1="arrow.x1"
        :y1="-arrow.y1"
        :x2="arrow.endX"
        :y2="arrow.endY"
        :stroke="arrow.color"
        :stroke-width="0.006"
        stroke-opacity="0.8"
        :marker-end="arrow.marker"
      />
      <SvgText :x="arrow.endX" :y="-arrow.endY" :text="arrow.speedText" :color="arrow.color" />
    </g>
  </g>
</template>
