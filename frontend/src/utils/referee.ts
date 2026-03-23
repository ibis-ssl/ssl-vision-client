import type { ToastLevel } from '@/composables/toast'

export const REFEREE_COMMAND_LABELS: Record<number, string> = {
  0: 'HALT',
  1: 'STOP',
  2: 'NORMAL_START',
  3: 'FORCE_START',
  4: 'PREPARE_KICKOFF_YELLOW',
  5: 'PREPARE_KICKOFF_BLUE',
  6: 'PREPARE_PENALTY_YELLOW',
  7: 'PREPARE_PENALTY_BLUE',
  8: 'DIRECT_FREE_YELLOW',
  9: 'DIRECT_FREE_BLUE',
  10: 'INDIRECT_FREE_YELLOW',
  11: 'INDIRECT_FREE_BLUE',
  12: 'TIMEOUT_YELLOW',
  13: 'TIMEOUT_BLUE',
  14: 'GOAL_YELLOW',
  15: 'GOAL_BLUE',
  16: 'BALL_PLACEMENT_YELLOW',
  17: 'BALL_PLACEMENT_BLUE',
}

const URGENT_COMMANDS = new Set([0, 1, 14, 15]) // HALT, STOP, GOAL_YELLOW, GOAL_BLUE

export function isUrgentCommand(cmd: number): boolean {
  return URGENT_COMMANDS.has(cmd)
}

export interface GameEventInfo {
  label: string
  level: ToastLevel
}

export const GAME_EVENT_INFO: Record<string, GameEventInfo> = {
  // ゴール
  goal:                              { label: 'ゴール',                    level: 'success' },
  possibleGoal:                      { label: 'ゴール (確認中)',             level: 'success' },
  invalidGoal:                       { label: '無効ゴール',                 level: 'warning' },
  // ボールアウト
  ballLeftFieldTouchLine:            { label: 'ボールアウト (サイド)',        level: 'info'    },
  ballLeftFieldGoalLine:             { label: 'ボールアウト (ゴールライン)',   level: 'info'    },
  aimlessKick:                       { label: 'エイムレスキック',            level: 'warning' },
  // ファール
  attackerTooCloseToDefenseArea:     { label: '攻撃側接近 (DF エリア)',      level: 'warning' },
  defenderInDefenseArea:             { label: 'DF エリア内 (GK 以外)',       level: 'warning' },
  boundaryCrossing:                  { label: 'ライン越え',                  level: 'warning' },
  keeperHeldBall:                    { label: 'GK ボール保持',              level: 'warning' },
  botDribbledBallTooFar:             { label: 'ドリブル超過',                level: 'warning' },
  botPushedBot:                      { label: '押し出し',                   level: 'warning' },
  botHeldBallDeliberately:           { label: 'ボール保持',                 level: 'warning' },
  botTippedOver:                     { label: '転倒',                       level: 'warning' },
  botDroppedParts:                   { label: '部品落下',                   level: 'warning' },
  attackerTouchedBallInDefenseArea:  { label: '攻撃側 DF エリア内接触',      level: 'warning' },
  botKickedBallTooFast:              { label: 'キック速度超過',              level: 'warning' },
  botCrashUnique:                    { label: 'クラッシュ',                 level: 'warning' },
  botCrashDrawn:                     { label: 'クラッシュ (相互)',           level: 'warning' },
  defenderTooCloseToKickPoint:       { label: 'キック点接近',               level: 'warning' },
  botTooFastInStop:                  { label: 'STOP 中速度超過',            level: 'warning' },
  botInterferedPlacement:            { label: 'プレイスメント妨害',          level: 'warning' },
  attackerDoubleTouchedBall:         { label: 'ダブルタッチ',               level: 'warning' },
  // プレイスメント
  placementSucceeded:                { label: 'プレイスメント成功',          level: 'success' },
  placementFailed:                   { label: 'プレイスメント失敗',          level: 'warning' },
  // その他
  noProgressInGame:                  { label: '試合停滞',                   level: 'info'    },
  penaltyKickFailed:                 { label: 'ペナルティキック失敗',        level: 'info'    },
  multipleCards:                     { label: '複数カード',                  level: 'warning' },
  multipleFouls:                     { label: '複数ファール',               level: 'warning' },
  botSubstitution:                   { label: 'ロボット交代',               level: 'info'    },
  excessiveBotSubstitution:          { label: '交代超過',                   level: 'warning' },
  tooManyRobots:                     { label: 'ロボット数超過',              level: 'warning' },
  challengeFlag:                     { label: 'チャレンジフラグ',           level: 'info'    },
  challengeFlagHandled:              { label: 'チャレンジ処理済み',          level: 'info'    },
  emergencyStop:                     { label: '緊急停止',                   level: 'warning' },
  unsportingBehaviorMinor:           { label: '非スポーツ行為 (軽微)',       level: 'warning' },
  unsportingBehaviorMajor:           { label: '非スポーツ行為 (重大)',       level: 'error'   },
}

// Team enum: UNKNOWN=0, YELLOW=1, BLUE=2
export function teamLabel(team: number | undefined): string {
  if (team === 1) return 'Yellow'
  if (team === 2) return 'Blue'
  return ''
}
