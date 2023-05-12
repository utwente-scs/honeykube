CREATE DATABASE IF NOT EXISTS userdb;
USE userdb;

CREATE TABLE IF NOT EXISTS users
(
Username varchar(100) NOT NULL,
Password varchar (100) NOT NULL,
CreditCard varchar(100),
PRIMARY KEY(Username)
);