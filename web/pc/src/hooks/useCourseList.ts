import { useEffect, useState } from 'react';
import type { Course } from '../api/course';

export function useCourseList(loader: () => Promise<Course[]>) {
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    loader()
      .then((result) => {
        if (active) setCourses(result);
      })
      .catch(() => {
        if (active) setCourses([]);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [loader]);

  return { courses, loading };
}
