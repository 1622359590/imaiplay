import {
  ClockCircleFilled,
  PlaySquareFilled,
} from '@ant-design/icons';
import { Card } from 'antd';
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
  return (
    <section className="learning-summary" aria-label="学习概览">
      <Card className="learning-summary-card" bordered={false}>
        <div className="learning-summary-card-heading">
          <span className="learning-summary-icon learning-summary-icon-progress" aria-hidden="true">
            <PlaySquareFilled />
          </span>
          <h2>课程进度</h2>
        </div>
        <p className="learning-summary-value">
          <span>必修课：</span>
          <span>已学完课程 <strong>{completed}</strong> / {required}</span>
        </p>
      </Card>

      <Card className="learning-summary-card" bordered={false}>
        <div className="learning-summary-card-heading">
          <span className="learning-summary-icon learning-summary-icon-time" aria-hidden="true">
            <ClockCircleFilled />
          </span>
          <h2>学习时长</h2>
        </div>
        <p className="learning-summary-value learning-time-values">
          <span>今日： <strong>{learningMinutes(todaySeconds)}</strong> 分钟</span>
          <span>累计： <strong>{learningMinutes(totalSeconds)}</strong> 分钟</span>
        </p>
      </Card>
    </section>
  );
}
