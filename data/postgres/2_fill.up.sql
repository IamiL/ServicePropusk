CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
INSERT INTO
    buildings(id, name, description, status, img_url)
VALUES
    (uuid_generate_v4(), 'Главный корпус (ГУК)', 'Оформление пропуска в главный учебный корпус по адресу 2-я Бауманская улица, 5', true, '/buildings/0.png'),
    (uuid_generate_v4(), 'Учебно-лабораторный корпус', 'Оформление пропуска в учебно-лабораторный корпус по адресу 2-я Бауманская улица, 5', true, '/buildings/1.png'),
    (uuid_generate_v4(), 'Корпус Э', 'Оформление пропуска в корпус "энерго" по адресу 2-я Бауманская улица, 5', true, '/buildings/2.png'),
    (uuid_generate_v4(), 'Корпус СМ', 'Оформление пропуска в корпус "специальное машиностроение" по адресу 2-я Бауманская улица, 5', true, '/buildings/3.png'),
    (uuid_generate_v4(), 'Корпус Т', 'Оформление пропуска в корпус "т" по адресу 2-я Бауманская улица, 5', true, '/buildings/4.png');