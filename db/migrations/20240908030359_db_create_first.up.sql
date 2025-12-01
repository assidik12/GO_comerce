-- Active: 1722607424659@@127.0.0.1@3306@go_rest_api

CREATE TABLE `category` (
    `id` int NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `stock` INT NOT NULL,
    PRIMARY KEY (`id`)
) engine = InnoDB;

CREATE TABLE `users` (
    `id` int NOT NULL AUTO_INCREMENT,
    `username` varchar(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL,
    `role` VARCHAR(255) NOT NULL DEFAULT 'user',
    `password` VARCHAR(300) NOT NULL,
    PRIMARY KEY (`id`)
) engine = InnoDB;

CREATE TABLE `products` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `price` INT NOT NULL,
    `stock` INT NOT NULL,
    `img` VARCHAR(255) NOT NULL,
    `description` VARCHAR(255) NOT NULL,
    `category_id` INT NOT NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`category_id`) REFERENCES `category` (`id`)
) ENGINE = InnoDB;

CREATE TABLE `transactions` (
    `id` VARCHAR(300) NOT NULL PRIMARY KEY,
    `user_id` INT NOT NULL,
    `created_at` DATETIME NOT NULL,
    `total_price` INT NOT NULL,
    Foreign Key (`user_id`) REFERENCES `users` (`id`)
) engine = InnoDB;

CREATE TABLE `transaction_details` (
    `transaction_id` VARCHAR(300) NOT NULL,
    `product_id` INT NOT NULL,
    `price` INT NOT NULL,
    `quantity` INT NOT NULL,
    Foreign Key (`transaction_id`) REFERENCES `transactions` (`id`),
    Foreign Key (`product_id`) REFERENCES `products` (`id`)
) engine = InnoDB;