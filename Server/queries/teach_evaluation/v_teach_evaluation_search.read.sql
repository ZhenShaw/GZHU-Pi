-- 搜索框模糊查询
select *
from v_teach_evaluation
where course_name LIKE '%{{.match}}%'
   OR teacher LIKE '%{{.match}}%'
   OR course_type LIKE '%{{.match}}%'
   OR constitute LIKE '%{{.match}}%'
   OR college LIKE '%{{.match}}%'
   OR title LIKE '%{{.match}}%'
   OR education LIKE '%{{.match}}%'
ORDER BY created_at DESC
LIMIT 30
