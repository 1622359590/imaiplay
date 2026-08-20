import {
  acknowledgeAndContinue,
  formatLearningDuration,
  motivationTargetPath,
  type LearnerMotivation,
} from '@imaiplay/shared/learning/learnerMotivation'
import { Button, Popup } from 'antd-mobile'
import {
  CheckCircleFill,
  ClockCircleOutline,
  HistogramOutline,
  PlayOutline,
  RightOutline,
} from 'antd-mobile-icons'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  acknowledgeLearnerMotivation,
  getLearnerMotivation,
} from '../api/motivation'
import { useTenantTheme } from '../context/TenantThemeContext'

interface LearnerMotivationPromptProps {
  enabled: boolean
}

function DailyResults({ prompt }: { prompt: Extract<LearnerMotivation, { kind: 'daily_summary' }> }) {
  const comparison = prompt.comparison?.exceededPercent !== undefined
    ? `超过企业内 ${prompt.comparison.exceededPercent}% 的活跃学员`
    : prompt.comparison?.durationChangeSeconds !== undefined && prompt.comparison.durationChangeSeconds > 0
      ? `比前一天多学习 ${formatLearningDuration(prompt.comparison.durationChangeSeconds)}`
      : undefined
  return (
    <>
      <div className="learner-motivation-metrics" aria-label="昨日学习成果">
        <div className="learner-motivation-metric"><ClockCircleOutline /><span>学习时长</span><strong>{formatLearningDuration(prompt.metrics.yesterdaySeconds)}</strong></div>
        <div className="learner-motivation-metric"><HistogramOutline /><span>学习课时</span><strong>{prompt.metrics.lessonCount} 个</strong></div>
        <div className="learner-motivation-metric"><CheckCircleFill /><span>完成课时</span><strong>{prompt.metrics.completedLessonCount} 个</strong></div>
        <div className="learner-motivation-metric"><CheckCircleFill /><span>必修进度</span><strong>{prompt.metrics.requiredCompleted}/{prompt.metrics.requiredTotal}</strong></div>
      </div>
      {comparison && <p className="learner-motivation-comparison"><HistogramOutline /> {comparison}</p>}
    </>
  )
}

export function LearnerMotivationPrompt({ enabled }: LearnerMotivationPromptProps) {
  const navigate = useNavigate()
  const theme = useTenantTheme()
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

  const close = () => {
    setVisible(false)
    void acknowledgeOnce().catch(() => undefined)
  }
  const continueLearning = () => {
    const target = motivationTargetPath(prompt)
    if (!target) return
    void acknowledgeAndContinue(
      acknowledgeOnce,
      () => navigate(theme.routePath(target)),
    ).catch(() => undefined)
  }

  return (
    <Popup
      position="bottom"
      visible={visible}
      bodyClassName="learner-motivation-popup"
      showCloseButton
      closeOnMaskClick
      onClose={close}
      onMaskClick={close}
      afterShow={() => { void acknowledgeOnce().catch(() => undefined) }}
    >
      <section role="dialog" aria-modal="true" aria-labelledby="h5-motivation-title">
        <div className="learner-motivation-heading-icon" aria-hidden="true"><PlayOutline /></div>
        <span className="learner-motivation-eyebrow">
          {prompt.kind === 'welcome' ? '新的开始' : prompt.kind === 'daily_summary' ? '学习成果' : '继续前进'}
        </span>
        <h2 id="h5-motivation-title">{prompt.title}</h2>
        <p className="learner-motivation-message">{prompt.message}</p>
        {prompt.kind === 'daily_summary' && <DailyResults prompt={prompt} />}
        <div className="learner-motivation-course">
          <div><span>{prompt.course.progressPercent > 0 ? '继续学习' : '推荐开始'}</span><strong>{prompt.course.title}</strong><small>{prompt.course.lessonTitle}</small></div>
          <b>{prompt.course.progressPercent}%</b>
        </div>
        <Button className="learner-motivation-primary" color="primary" block onClick={continueLearning}>
          {prompt.course.progressPercent > 0 ? '继续学习' : '开始第一课'} <RightOutline />
        </Button>
        <button className="learner-motivation-later" type="button" onClick={close}>稍后再看</button>
      </section>
    </Popup>
  )
}
