CREATE DATABASE IF NOT EXISTS productcatalog;
USE productcatalog;

CREATE TABLE IF NOT EXISTS products
(
id varchar(100) NOT NULL,
name varchar (200) NOT NULL,
description text,
img_path varchar(200),
currency_code char(3),
units int,
nanos int,
categories varchar(500),
PRIMARY KEY(id)
);

INSERT INTO products VALUES
(
    "OLJCESPC7Z",
    "Vintage Typewriter",
    "This typewriter looks good in your living room.",
    "/static/img/products/typewriter.jpg",
    "USD",
    67,
    990000000,
    "vintage"
);

INSERT INTO products VALUES
(
    "66VCHSJNUP",
    "Vintage Camera Lens",
    "You won't have a camera to use it and it probably doesn't work anyway.",
    "/static/img/products/camera-lens.jpg",
    "USD",
    12,
    490000000,
    "photography,vintage"
);

INSERT INTO products VALUES
(
    "1YMWWN1N4O",
    "Home Barista Kit",
    "Always wanted to brew coffee with Chemex and Aeropress at home?",
    "/static/img/products/barista-kit.jpg",
    "USD",
    124,
    0,
    "cookware"
);

INSERT INTO products VALUES
(
    "L9ECAV7KIM",
    "Terrarium",
    "This terrarium will looks great in your white painted living room.",
    "/static/img/products/terrarium.jpg",
    "USD",
    36,
    450000000,
    "gardening"
);

INSERT INTO products VALUES
(
    "2ZYFJ3GM2N",
    "Film Camera",
    "This camera looks like it's a film camera, but it's actually digital.",
    "/static/img/products/film-camera.jpg",
    "USD",
    2245,
    0,
    "photography,vintage"
);

INSERT INTO products VALUES
(
    "0PUK6V6EV0",
    "Vintage Record Player",
    "It still works.",
    "/static/img/products/record-player.jpg",
    "USD",
    65,
    500000000,
    "music,vintage"
);

INSERT INTO products VALUES
(
    "LS4PSXUNUM",
    "Metal Camping Mug",
    "You probably don't go camping that often but this is better than plastic cups.",
    "/static/img/products/camp-mug.jpg",
    "USD",
    24,
    330000000,
    "cookware"
);

INSERT INTO products VALUES
(
    "9SIQT8TOJO",
    "City Bike",
    "This single gear bike probably cannot climb the hills of San Francisco.",
    "/static/img/products/city-bike.jpg",
    "USD",
    789,
    500000000,
    "cycling"
);

INSERT INTO products VALUES
(
    "6E92ZMYYFZ",
    "Air Plant",
    "Have you ever wondered whether air plants need water? Buy one and figure out.",
    "/static/img/products/air-plant.jpg",
    "USD",
    12,
    300000000,
    "gardening"
);
