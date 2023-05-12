# User data access service
import datetime
import logging
import os

import pymysql
from flask import Flask, jsonify, request, session
from flask_bcrypt import check_password_hash, generate_password_hash
from flaskext.mysql import MySQL

app = Flask(__name__)
app.secret_key = "honeykube"
logging.basicConfig(
    level=logging.DEBUG,
    format=f'%(asctime)s %(levelname)s %(name)s : %(message)s'
)

mysql = MySQL()

# MySQL configurations (Get from env variables later)
app.config['MYSQL_DATABASE_USER'] = os.environ["MYSQL_USER"]
app.config['MYSQL_DATABASE_PASSWORD'] = os.environ["MYSQL_PASSWORD"]
app.config['MYSQL_DATABASE_DB'] = os.environ["MYSQL_DB_NAME"]
app.config['MYSQL_DATABASE_HOST'] = os.environ["MYSQL_DB_SERVICE_ADDR"]
mysql.init_app(app)


@app.before_request
def before_request():
    """
        Called before every request to reset session lifetime
        so that session expires if user is inactive for more
        than defined time delta
    """
    session.permanent = True
    app.permanent_session_lifetime = datetime.timedelta(days=1)
    session.modified = True


@app.after_request
def after_request(response):
    """
        Called after every request
    """
    response.headers.add('Access-Control-Allow-Credentials', "true")
    return response


def create_response_message(status_code, msg):
    return_msg = {"status_code": status_code, "message": msg}
    return jsonify(return_msg), status_code


@app.route('/', methods=['GET'])
def index():
    return create_response_message(200, "Successfully reached userdbservice!")


@app.route('/register', methods=['POST'])
def register_user():
    try:
        user_data = request.json
        cursor = mysql.get_db().cursor()
        cursor.execute(
            ''' INSERT INTO users (Username, Password) VALUES(%s,%s)''',
            (
                user_data["username"],
                generate_password_hash(user_data["password"])
            )
        )
        mysql.get_db().commit()

        cursor.close()
        app.logger.info("Successfully registered user {}".format(
            user_data["username"]
            )
        )
        session["Username"] = user_data["username"]
        return create_response_message(200, "Successful Registration!")

    except pymysql.err.IntegrityError as e:
        app.logger.error(str(e))
        return create_response_message(403, str(e))

    except Exception as e:
        return create_response_message(401, str(e))


@app.route('/login', methods=['POST'])
def login():
    try:
        user_data = request.json
        cursor = mysql.get_db().cursor()
        sql_query = "SELECT Password FROM users WHERE Username='{}'".format(
                user_data["username"]
            )
        cursor.execute(sql_query)
        data = cursor.fetchone()
        mysql.get_db().commit()
        cursor.close()
        if data:
            if(check_password_hash(data[0], user_data["password"])):
                session["Username"] = user_data["username"]
                app.logger.info("Successfully logged in user: {}".format(
                        user_data["username"]
                    )
                )
                return create_response_message(200, "Successful Login!")
            else:
                app.logger.error("Invalid username or password.")
                return create_response_message(
                    400,
                    "Invalid username or password."
                )
        else:
            app.logger.error("Username {} or password doesn't exist!".format(
                    user_data["username"]
                )
            )
            return create_response_message(
                    400,
                    "Username or password doesn't exist!"
                )

    except Exception as e:
        return create_response_message(401, str(e))


@app.route('/logout', methods=['GET'])
def logout():
    session.pop('Username', None)
    app.logger.info("User logged out! Session Exited.")
    return create_response_message(200, "Logged Out")


@app.route('/add_credit_card', methods=['POST'])
def add_credit_card():
    try:
        user_data = request.json
        print(user_data)
        cursor = mysql.get_db().cursor()
        cursor.execute(
            ''' UPDATE users SET CreditCard = %s WHERE Username = %s''',
            (
                user_data["credit_card_number"],
                user_data["username"]
            )
        )
        mysql.get_db().commit()

        cursor.close()
        app.logger.info("Credit card number saved for the user {}".format(
                user_data["username"]
            )
        )
        return create_response_message(
            200,
            "Successful Saved Credit Card Number!"
        )

    except Exception as e:
        app.logger.error(str(e))
        return create_response_message(401, str(e))


@app.route('/get_user_info', methods=['POST'])
def get_user_info():
    try:
        user_data = request.json
        print(user_data)
        cursor = mysql.get_db().cursor()
        sql_query = "SELECT CreditCard FROM users WHERE Username = '{}'".format(
                user_data["username"]
            )
        cursor.execute(sql_query)
        data = cursor.fetchone()
        mysql.get_db().commit()
        cursor.close()
        return_data = {"status_code": 200}
        return_data["username"] = user_data["username"]

        if data:
            return_data["credit_card"] = data[0]
        else:
            return_data["credit_card"] = ""

        app.logger.info("Credit Card information requested by user {}".format(
                user_data["username"]
            )
        )
        return return_data

    except Exception as e:
        app.logger.error(str(e))
        return create_response_message(401, str(e))


if __name__ == '__main__':
    app.run(port=5000, host='0.0.0.0', debug=True)
