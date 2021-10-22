.PHONY: compilemessages
compilemessages:
	python manage.py compilemessages --settings=backend.settings.ce.dev

.PHONY: makemessages
makemessages:
	python manage.py makemessages -l zh-hans -l en -i "assets" -i "frontend" --settings=backend.settings.ce.dev

.PHONY: requirements.txt
requirements.txt:
	# 去掉后面的 python 版本, 支持老版本的 pip 安装, 兼容标准出包流程
	poetry export -f requirements.txt --without-hashes|cut -f1 -d";" > requirements.txt && \
	sed -i '' '/colorama/d' requirements.txt && \
	sed -i '' '/twisted-iocpsupport/d' requirements.txt
