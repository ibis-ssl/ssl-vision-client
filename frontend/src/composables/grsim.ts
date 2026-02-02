/**
 * grSim API composable
 */

const API_BASE = '/api/grsim';

export interface BallReplacementRequest {
  x: number;  // メートル単位
  y: number;  // メートル単位
}

export interface RobotReplacementRequest {
  x: number;          // メートル単位
  y: number;          // メートル単位
  dir: number;        // 度単位
  id: number;         // ロボットID
  yellowTeam: boolean;// true=黄色チーム、false=青色チーム
}

export function useGrSimReplacement() {
  /**
   * ボールを指定位置に配置
   */
  const replaceBall = async (x: number, y: number): Promise<void> => {
    try {
      const response = await fetch(`${API_BASE}/ball`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ x, y } as BallReplacementRequest),
      });

      if (!response.ok) {
        throw new Error(`Failed to replace ball: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error replacing ball:', error);
      throw error;
    }
  };

  /**
   * ロボットを指定位置に配置
   */
  const replaceRobot = async (
    x: number,
    y: number,
    dir: number,
    id: number,
    yellowTeam: boolean
  ): Promise<void> => {
    try {
      const requestData = { x, y, dir, id, yellowTeam } as RobotReplacementRequest;

      const response = await fetch(`${API_BASE}/robot`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
      });

      if (!response.ok) {
        throw new Error(`Failed to replace robot: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Error replacing robot:', error);
      throw error;
    }
  };

  return {
    replaceBall,
    replaceRobot,
  };
}
