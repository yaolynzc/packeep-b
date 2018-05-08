# 基于Go实现的纯后台api服务
## 对应前端详见仓库：packeep-f
## 采用框架如下：
## gorm + httprouter
## 问题记录：
### 1、定义模型时即使给字段设置了默认值，gorm执行update操作时，必须每个字段赋值才会成功，否则更新无效。
### 2、gorm执行distinct查询，find方法每次只返回1条数据，必须采用sql.rows()方法查询，并循环rows自定保存结果
