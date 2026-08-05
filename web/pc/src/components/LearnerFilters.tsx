import { Select, Tabs } from 'antd';
import type { LearnerCategory } from '../api/learner';
import type { LearnerCourseTab } from '../utils/learnerCourses';

const tabItems: Array<{ key: LearnerCourseTab; label: string }> = [
  { key: 'all', label: '全部' },
  { key: 'required', label: '必修课' },
  { key: 'optional', label: '选修课' },
  { key: 'completed', label: '已学完' },
  { key: 'incomplete', label: '未学完' },
];

interface LearnerFiltersProps {
  tab: LearnerCourseTab;
  categoryId?: string;
  categories: LearnerCategory[];
  onTabChange: (tab: LearnerCourseTab) => void;
  onCategoryChange: (categoryId?: string) => void;
}

export function LearnerFilters({
  tab,
  categoryId,
  categories,
  onTabChange,
  onCategoryChange,
}: LearnerFiltersProps) {
  return (
    <div className="learner-filters">
      <Tabs
        className="learner-filter-tabs"
        activeKey={tab}
        items={tabItems}
        onChange={(key) => onTabChange(key as LearnerCourseTab)}
      />
      <Select
        aria-label="课程分类"
        className="learner-category-select"
        value={categoryId ?? 'all'}
        options={[
          { value: 'all', label: '所有分类' },
          ...categories.map((category) => ({ value: category.id, label: category.name })),
        ]}
        onChange={(value) => onCategoryChange(value === 'all' ? undefined : value)}
      />
    </div>
  );
}
