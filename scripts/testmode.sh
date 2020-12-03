# testmode.sh is used to make :victorrds/etebase suitable to testing.
# It enables the user signup via API
sed -e '/ETEBASE_CREATE_USER_FUNC/ s/^#*/#/' -i etebase_server/settings.py

./manage.py migrate
./manage.py runserver 0.0.0.0:3735
