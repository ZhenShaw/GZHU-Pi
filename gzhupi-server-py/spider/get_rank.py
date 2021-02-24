from models.models import *


class SQL(object):

    def __init__(self):
        # 查询大学GPA排名
        self.gpa_rank = """
            select *
            from (SELECT stu_id,gpa,colllage_rank,major_rank,
                         ROW_NUMBER()
                         over (order by gpa desc) as class_rank
                  FROM (SELECT class_id, stu_id, gpa, colllage_rank, ROW_NUMBER()
                               over (order by gpa desc) as major_rank
                        FROM (SELECT major_id, class_id, stu_id, gpa, ROW_NUMBER()
                                     over (order by gpa desc) as colllage_rank
                              FROM (select major_id, class_id, stu_id,
                                           cast(sum(credit * course_gpa) / sum(credit) AS decimal(3, 2)) as gpa
                                    from v_grade
                                    where stu_id in (select distinct stu_id --学院同级所有学生
                                                     from v_grade
                                                     where college_id = (select college_id from t_stu_info where stu_id = '%s')
                                                       and admit_year = (select admit_year from t_stu_info where stu_id = '%s'))
                                      and course_gpa != 0
                                    group by major_id, class_id, stu_id) as a) as b
                        where major_id = (select distinct major_id from t_stu_info where stu_id = '%s')) as c
                  where class_id = (select distinct class_id from t_stu_info where stu_id = '%s')) as d
            where stu_id = '%s';
        """

        self.year_rank = """
            select * from (SELECT year, stu_id, gpa, colllage_rank, major_rank,
                         ROW_NUMBER() over (PARTITION By year order by gpa desc) as class_rank
                  FROM (SELECT year, class_id, stu_id, gpa, colllage_rank,
                               ROW_NUMBER() over (PARTITION By year order by gpa desc) as major_rank
                        FROM (SELECT year, major_id, class_id, stu_id, gpa,
                                     ROW_NUMBER() over (PARTITION By year order by gpa desc) as colllage_rank
                              FROM (select year, major_id, class_id, stu_id,
                                           cast(sum(credit * course_gpa) / sum(credit) AS decimal(3, 2)) as gpa
                                    from v_grade
                                    where stu_id in (select distinct stu_id from v_grade
                                                     where college_id = (select college_id from t_stu_info where stu_id = '%s')
                                                       and admit_year = (select admit_year from t_stu_info where stu_id = '%s'))
                                        and course_gpa != 0
                                    group by year, major_id, class_id, stu_id) as v) as a
                        where major_id = (select major_id from t_stu_info where stu_id = '%s')) as b
                  where class_id = (select class_id from t_stu_info where stu_id = '%s')
                  order by stu_id) as c
            where stu_id = '%s'
            order by year desc;
        """

        # 查询每学期GPA排名
        self.sem_rank = """
            select *
            from (SELECT year_sem,stu_id,gpa,colllage_rank,major_rank,
                         ROW_NUMBER()
                         over (PARTITION By year_sem order by gpa desc) as class_rank
                  FROM (SELECT year_sem, class_id, stu_id, gpa, colllage_rank,
                               ROW_NUMBER()
                               over (PARTITION By year_sem order by gpa desc) as major_rank
                        FROM (SELECT year_sem, major_id, class_id, stu_id, gpa,
                                     ROW_NUMBER()
                                     over (PARTITION By year_sem order by gpa desc) as colllage_rank
                              FROM (select year_sem, major_id, class_id, stu_id,
                                           cast(sum(credit * course_gpa) / sum(credit) AS decimal(3, 2)) as gpa
                                    from v_grade
                                    where stu_id in (select distinct stu_id --学院同级所有学生
                                                     from v_grade 
                                                     where college_id = (select college_id from t_stu_info where stu_id = '%s')
                                                       and admit_year = (select admit_year from t_stu_info where stu_id = '%s'))
                                        and invalid != '是'
                                    group by year_sem, major_id, class_id, stu_id) as v) as a
                        where major_id = (select major_id from t_stu_info where stu_id = '%s')) as b
                  where class_id = (select class_id from t_stu_info where stu_id = '%s')
                  order by stu_id) as c
            where stu_id = '%s'
            order by year_sem desc;
        """

        # 一键查询某学号的所有科目成绩及其排名
        self.course_rank = """
            select year_sem, course_name, course_id, grade_value, college_rank, major_rank, class_rank
            from (select stu_id, course_name, course_id, grade_value, year_sem, college_rank, major_rank,
                         ROW_NUMBER()
                             over
                               (PARTITION By course_id
                               order by grade_value desc) as class_rank
                  from (select stu_id, class_id, course_name, course_id, grade_value, year_sem, college_rank,
                               ROW_NUMBER()
                                   over
                                     (PARTITION By course_id
                                     order by grade_value desc) as major_rank
                        from (SELECT stu_id, major_id, class_id, course_name, course_id, grade_value, year_sem,
                                     ROW_NUMBER()
                                         over
                                           (PARTITION By course_id
                                           order by grade_value desc) as college_rank
                              FROM v_grade,
                                   (select distinct college_id, admit_year from t_stu_info where stu_id = '%s') as tmp
                              where v_grade.college_id = tmp.college_id
                                and v_grade.admit_year = tmp.admit_year) as a
                        where major_id = (select distinct major_id from t_stu_info where stu_id = '%s')) as b
                  where class_id = (select distinct class_id from t_stu_info where stu_id = '%s')) as c
            where stu_id = '%s'
            order by year_sem desc;
        """

        # 统计人数
        self.stu_count = """
            select (select count(*) from t_stu_info where college_id = a.college_id and admit_year = a.admit_year) as college_count,
           (select count(*) from t_stu_info where major_id = a.major_id and admit_year = a.admit_year)     as major_count,
           (select count(*) from t_stu_info where class_id = a.class_id and admit_year = a.admit_year)     as class_count
            from (select * from t_stu_info where stu_id = '%s') as a;
        """


m = None

s = SQL()


def get_gpa_rank(username):
    global m
    if m is None:
        m = Models()
    query = s.gpa_rank % (username, username, username, username, username)
    rows = m.query(query)
    gpa_rank = {}
    if len(rows) > 0:
        gpa_rank["stu_id"] = rows[0][0]
        gpa_rank["gpa"] = "%s" % rows[0][1]
        gpa_rank["college_rank"] = rows[0][2]
        gpa_rank["major_rank"] = rows[0][3]
        gpa_rank["class_rank"] = rows[0][4]

    return gpa_rank


def get_year_rank(username):
    global m
    if m is None:
        m = Models()
    query = s.year_rank % (username, username, username, username, username)
    rows = m.query(query)
    year_rank = []
    for item in rows:
        temp = {}
        temp["year"] = item[0]
        temp["stu_id"] = item[1]
        temp["gpa"] = "%s" % item[2]
        temp["college_rank"] = item[3]
        temp["major_rank"] = item[4]
        temp["class_rank"] = item[5]
        year_rank.append(temp)

    return year_rank


def get_sem_rank(username):
    global m
    if m is None:
        m = Models()
    query = s.sem_rank % (username, username, username, username, username)
    rows = m.query(query)
    sem_rank = []
    for item in rows:
        temp = {}
        temp["year_sem"] = item[0]
        temp["stu_id"] = item[1]
        temp["gpa"] = "%s" % item[2]
        temp["college_rank"] = item[3]
        temp["major_rank"] = item[4]
        temp["class_rank"] = item[5]
        sem_rank.append(temp)

    return sem_rank


def get_course_rank(username):
    global m
    if m is None:
        m = Models()
    query = s.course_rank % (username, username, username, username)
    rows = m.query(query)
    course_rank = []
    for item in rows:
        temp = {}
        temp["year_sem"] = item[0]
        temp["course_name"] = item[1]
        temp["course_id"] = item[2]
        temp["grade_value"] = item[3]
        temp["college_rank"] = item[4]
        temp["major_rank"] = item[5]
        temp["class_rank"] = item[6]
        course_rank.append(temp)

    return course_rank


def get_stu_count(username):
    global m
    if m is None:
        m = Models()
    query = s.stu_count % (username)
    rows = m.query(query)
    print(rows)
    stu_count = {}
    if len(rows) > 0:
        stu_count["college_count"] = rows[0][0]
        stu_count["major_count"] = rows[0][1]
        stu_count["class_count"] = rows[0][2]

    return stu_count


def get_rank(username):
    global m
    if m is None:
        m = Models()
    rank = {
        "stu_count": get_stu_count(username),
        "gpa_rank": get_gpa_rank(username),
        "year_rank": get_year_rank(username),
        "sem_rank": get_sem_rank(username),
        "course_rank": get_course_rank(username)
    }

    return rank
