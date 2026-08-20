import {
  ArrowRightOutlined,
  BookOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  PlayCircleFilled,
  RiseOutlined,
  TrophyFilled,
} from '@ant-design/icons'
import {
  acknowledgeAndContinue,
  formatLearningDuration,
  motivationTargetPath,
  type LearnerMotivation,
} from '@imaiplay/shared/learning/learnerMotivation'
import { Button, Modal } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  acknowledgeLearnerMotivation,
  getLearnerMotivation,
} from '../api/motivation'
import { usePortal } from '../context/PortalContext'
import { portalRoutePath } from '../utils/portalRouting'

interface LearnerMotivationPromptProps {
  enabled: boolean
}

function Comparison({ prompt }: { prompt: Extract<LearnerMotivation, { kind: 'daily_summary' }> }) {
  const exceeded = prompt.comparison?.exceededPercent
  const change = prompt.comparison?.durationChangeSeconds
  if (exceeded !== undefined) {
    return <p className="learner-motivation-comparison"><RiseOutlined /> 昨天的学习投入超过了企业内 {exceeded}% 的活跃学员</p>
  }
  if (change !== undefined && change > 0) {
    return <p className="learner-motivation-comparison"><RiseOutlined /> 比前一天多学习 {formatLearningDuration(change)}</p>
  }
  return null
}

function DailyMetrics({ prompt }: { prompt: Extract<LearnerMotivation, { kind: 'daily_summary' }> }) {
  const requiredProgress = prompt.metrics.requiredTotal > 0
    ? `${prompt.metrics.requiredCompleted}/${prompt.metrics.requiredTotal}`
    : '暂无'
  return (
    <div className="learner-motivation-metrics" aria-label="昨日学习成果">
      <div className="learner-motivation-metric"><ClockCircleOutlined /><span>学习时长</span><strong>{formatLearningDuration(prompt.metrics.yesterdaySeconds)}</strong></div>
      <div className="learner-motivation-metric"><BookOutlined /><span>学习课时</span><strong>{prompt.metrics.lessonCount} 个</strong></div>
      <div className="learner-motivation-metric"><CheckCircleFilled /><span>完成课时</span><strong>{prompt.metrics.completedLessonCount} 个</strong></div>
      <div className="learner-motivation-metric"><TrophyFilled /><span>必修进度</span><strong>{requiredProgress}</strong></div>
    </div>
  )
}

export function LearnerMotivationPrompt({ enabled }: LearnerMotivationPromptProps) {
  const navigate = useNavigate()
  const { mode, tenantCode } = usePortal()
  const [prompt, setPrompt] = useState<LearnerMotivation>({ kind: 'none' })
  const [visible, setVisible] = useState(false)
  const requestRef = useRef<Promise<LearnerMotivation>>()
  const acknowledgementRef = useRef<Promise<void>>()

  useEffect(() => {
    if (!enabled) {
      requestRef.current = undefined
      setPrompt({ kind: 'none' })
      setVisible(false)
      return
    }
    requestRef.current ??= getLearnerMotivation()
    let active = true
    void requestRef.current
      .then((result) => {
        if (!active) return
        setPrompt(result)
        setVisible(result.kind !== 'none')
      })
      .catch(() => undefined)
    return () => { active = false }
  }, [enabled])

  useEffect(() => {
    acknowledgementRef.current = undefined
  }, [prompt.kind === 'none' ? '' : prompt.promptKey])

  const acknowledgeOnce = useCallback(() => {
    if (prompt.kind === 'none') return Promise.resolve()
    acknowledgementRef.current ??= acknowledgeLearnerMotivation(prompt.promptKey)
    return acknowledgementRef.current
  }, [prompt])

  if (prompt.kind === 'none') return null

  const target = motivationTargetPath(prompt)
  const continueLearning = () => {
    if (!target) return
    const destination = portalRoutePath(mode, tenantCode, target)
    void acknowledgeAndContinue(acknowledgeOnce, () => navigate(destination))
      .catch(() => undefined)
  }

  return (
    <Modal
      className={`learner-motivation-modal learner-motivation-${prompt.kind}`}
      open={visible}
      width={560}
      footer={null}
      centered
      closable
      aria-labelledby="learner-motivation-title"
      onCancel={() => setVisible(false)}
      afterOpenChange={(open) => {
        if (open) void acknowledgeOnce().catch(() => undefined)
      }}
    >
      <div className="learner-motivation-heading-icon" aria-hidden="true">
        {prompt.kind === 'daily_summary' ? <TrophyFilled /> : <PlayCircleFilled />}
      </div>
      <p className="learner-motivation-eyebrow">
        {prompt.kind === 'welcome' ? '新的开始' : prompt.kind === 'daily_summary' ? '学习成果' : '继续前进'}
      </p>
      <h2 id="learner-motivation-title">{prompt.title}</h2>
      <p className="learner-motivation-message">{prompt.message}</p>
      {prompt.kind === 'daily_summary' && <DailyMetrics prompt={prompt} />}
      {prompt.kind === 'daily_summary' && <Comparison prompt={prompt} />}
      <div className="learner-motivation-course">
        <div><span>{prompt.course.progressPercent > 0 ? '继续学习' : '推荐开始'}</span><strong>{prompt.course.title}</strong><small>{prompt.course.lessonTitle}</small></div>
        <span>{prompt.course.progressPercent}%</span>
      </div>
      <div className="learner-motivation-actions">
        <Button type="text" onClick={() => setVisible(false)}>稍后再看</Button>
        <Button className="learner-motivation-primary" type="primary" onClick={continueLearning}>
          {prompt.course.progressPercent > 0 ? '继续学习' : '开始第一课'} <ArrowRightOutlined />
        </Button>
      </div>
    </Modal>
  )
}
