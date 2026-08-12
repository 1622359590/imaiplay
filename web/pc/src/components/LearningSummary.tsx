import {
  CheckCircleFilled,
  ClockCircleFilled,
  FireFilled,
  PlaySquareFilled,
} from '@ant-design/icons';
import { learningMinutes } from '../utils/learnerCourses';

interface LearningSummaryProps {
  completed: number;
  required: number;
  todaySeconds: number;
  totalSeconds: number;
}

export function LearningSummary({
  completed,
  required,
  todaySeconds,
  totalSeconds,
}: LearningSummaryProps) {
  const todayMinutes = learningMinutes(todaySeconds);
  const totalMinutes = learningMinutes(totalSeconds);
  const items = [
    { key: 'progress', icon: <PlaySquareFilled />, label: '进行中', value: Math.max(required - completed, 0), unit: '门课程', trend: `必修课程共 ${required} 门` },
    { key: 'complete', icon: <CheckCircleFilled />, label: '已完成', value: completed, unit: '门课程', trend: required ? `完成率 ${Math.round((completed / required) * 100)}%` : '等待课程安排' },
    { key: 'today', icon: <ClockCircleFilled />, label: '今日学习', value: todayMinutes, unit: '分钟', trend: todayMinutes ? '保持今天的学习节奏' : '开始今天的学习' },
    { key: 'total', icon: <FireFilled />, label: '累计学习', value: totalMinutes, unit: '分钟', trend: '每一分钟都算数' },
  ];
  return (
    <section className="learning-summary" aria-label="学习概览">
      {items.map((item) => (
        <article className={`learning-summary-card learning-summary-${item.key}`} key={item.key}>
          <span className="learning-summary-icon" aria-hidden="true">{item.icon}</span>
          <span className="learning-summary-label">{item.label}</span>
          <p className="learning-summary-value"><strong>{item.value}</strong><span>{item.unit}</span></p>
          <small>{item.trend}</small>
        </article>
      ))}
    </section>
  );
}
