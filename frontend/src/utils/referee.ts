import type { ToastLevel } from '@/composables/toast'
import { Referee_Command } from '@/proto/gc/ssl_gc_referee_message_pb.ts'

export const REFEREE_COMMAND_LABELS: Record<number, string> = {
  [Referee_Command.HALT]:                    'HALT',
  [Referee_Command.STOP]:                    'STOP',
  [Referee_Command.NORMAL_START]:            'NORMAL_START',
  [Referee_Command.FORCE_START]:             'FORCE_START',
  [Referee_Command.PREPARE_KICKOFF_YELLOW]:  'PREPARE_KICKOFF_YELLOW',
  [Referee_Command.PREPARE_KICKOFF_BLUE]:    'PREPARE_KICKOFF_BLUE',
  [Referee_Command.PREPARE_PENALTY_YELLOW]:  'PREPARE_PENALTY_YELLOW',
  [Referee_Command.PREPARE_PENALTY_BLUE]:    'PREPARE_PENALTY_BLUE',
  [Referee_Command.DIRECT_FREE_YELLOW]:      'DIRECT_FREE_YELLOW',
  [Referee_Command.DIRECT_FREE_BLUE]:        'DIRECT_FREE_BLUE',
  [Referee_Command.INDIRECT_FREE_YELLOW]:    'INDIRECT_FREE_YELLOW',
  [Referee_Command.INDIRECT_FREE_BLUE]:      'INDIRECT_FREE_BLUE',
  [Referee_Command.TIMEOUT_YELLOW]:          'TIMEOUT_YELLOW',
  [Referee_Command.TIMEOUT_BLUE]:            'TIMEOUT_BLUE',
  [Referee_Command.GOAL_YELLOW]:             'GOAL_YELLOW',
  [Referee_Command.GOAL_BLUE]:               'GOAL_BLUE',
  [Referee_Command.BALL_PLACEMENT_YELLOW]:   'BALL_PLACEMENT_YELLOW',
  [Referee_Command.BALL_PLACEMENT_BLUE]:     'BALL_PLACEMENT_BLUE',
}

const URGENT_COMMANDS = new Set([
  Referee_Command.HALT,
  Referee_Command.STOP,
  Referee_Command.GOAL_YELLOW,
  Referee_Command.GOAL_BLUE,
])

export function isUrgentCommand(cmd: number): boolean {
  return URGENT_COMMANDS.has(cmd)
}

export interface GameEventInfo {
  label: string
  level: ToastLevel
}

export const GAME_EVENT_INFO: Record<string, GameEventInfo> = {
  // ゴール
  goal:                              { label: 'ゴール',                        level: 'success' },
  possibleGoal:                      { label: 'ゴール (確認中)',                level: 'success' },
  invalidGoal:                       { label: '無効ゴール',                     level: 'warning' },
  // ボールアウト
  ballLeftFieldTouchLine:            { label: 'タッチライン通過',               level: 'info'    },
  ballLeftFieldGoalLine:             { label: 'ゴールライン通過',               level: 'info'    },
  aimlessKick:                       { label: 'エイムレスキック',               level: 'warning' },
  // 試合停止を伴うファール
  attackerTooCloseToDefenseArea:     { label: 'エリア接近',                     level: 'warning' },
  defenderInDefenseArea:             { label: 'マルチプルディフェンス',          level: 'warning' },
  boundaryCrossing:                  { label: '境界線横断',                     level: 'warning' },
  keeperHeldBall:                    { label: 'キーパーボール保持',             level: 'warning' },
  botDribbledBallTooFar:             { label: 'オーバードリブル',               level: 'warning' },
  botPushedBot:                      { label: 'プッシング',                     level: 'warning' },
  botHeldBallDeliberately:           { label: 'ホールディング',                 level: 'warning' },
  botTippedOver:                     { label: '転倒',                           level: 'warning' },
  botDroppedParts:                   { label: '部品脱落',                       level: 'warning' },
  // 試合停止を伴わないファール
  attackerTouchedBallInDefenseArea:  { label: 'エリア内タッチ',                 level: 'warning' },
  botKickedBallTooFast:              { label: 'ボール速度超過',                 level: 'warning' },
  botCrashUnique:                    { label: '衝突',                           level: 'warning' },
  botCrashDrawn:                     { label: '衝突 (引分)',                    level: 'warning' },
  // アウトオブプレイ中のファール
  defenderTooCloseToKickPoint:       { label: 'ボール接近',                     level: 'warning' },
  botTooFastInStop:                  { label: 'ストップ中速度超過',             level: 'warning' },
  botInterferedPlacement:            { label: '配置妨害',                       level: 'warning' },
  attackerDoubleTouchedBall:         { label: 'ダブルタッチ',                   level: 'warning' },
  // ボール配置
  placementSucceeded:                { label: 'ボール配置成功',                 level: 'success' },
  placementFailed:                   { label: 'ボール配置失敗',                 level: 'warning' },
  // その他
  noProgressInGame:                  { label: '試合の停滞',                     level: 'info'    },
  penaltyKickFailed:                 { label: 'ペナルティーキック失敗',          level: 'info'    },
  multipleCards:                     { label: '複数カード',                     level: 'warning' },
  multipleFouls:                     { label: '複数ファール → イエローカード',  level: 'warning' },
  botSubstitution:                   { label: 'ロボット交代',                   level: 'info'    },
  excessiveBotSubstitution:          { label: 'ロボット交代回数超過',           level: 'warning' },
  tooManyRobots:                     { label: 'ロボット数超過',                 level: 'warning' },
  challengeFlag:                     { label: 'チャレンジフラグ',               level: 'info'    },
  challengeFlagHandled:              { label: 'チャレンジ処理済み',             level: 'info'    },
  emergencyStop:                     { label: '緊急停止',                       level: 'warning' },
  unsportingBehaviorMinor:           { label: '非スポーツマン行為 (軽微)',       level: 'warning' },
  unsportingBehaviorMajor:           { label: '非スポーツマン行為 (重大)',       level: 'error'   },
  // 非推奨イベント（旧バージョン互換）
  prepared:                                        { label: '準備完了',                       level: 'info'    },
  indirectGoal:                                    { label: '間接ゴール',                     level: 'success' },
  chippedGoal:                                     { label: 'チップゴール',                   level: 'success' },
  kickTimeout:                                     { label: 'キックタイムアウト',             level: 'warning' },
  attackerTouchedOpponentInDefenseArea:             { label: 'エリア内接触',                   level: 'warning' },
  attackerTouchedOpponentInDefenseAreaSkipped:      { label: 'エリア内接触 (スキップ)',        level: 'info'    },
  botCrashUniqueSkipped:                           { label: '衝突 (スキップ)',                level: 'info'    },
  botPushedBotSkipped:                             { label: 'プッシング (スキップ)',          level: 'info'    },
  defenderInDefenseAreaPartially:                  { label: '部分的エリア侵入',               level: 'warning' },
  multiplePlacementFailures:                       { label: 'ボール配置連続失敗',             level: 'warning' },
}

// Team enum: UNKNOWN=0, YELLOW=1, BLUE=2
export function teamLabel(team: number | undefined): string {
  if (team === 1) return 'Yellow'
  if (team === 2) return 'Blue'
  return ''
}
