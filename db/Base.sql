CREATE DATABASE IF NOT EXISTS `ITmiraiApp`;

USE `ITmiraiApp`;

CREATE TABLE IF NOT EXISTS `ChangeOfClass` (
  `Date` date NOT NULL,
  `ClassID` varchar(5) NOT NULL,
  `Time` varchar(10) NOT NULL,
  `After` text NOT NULL,
  `Label` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `Users` (
  `uid` VARCHAR(100) NOT NULL,
  `name` varchar(10) NOT NULL,
  `email` text NOT NULL,
  `photoURL` text NOT NULL,
  `GradeInSchool` text NOT NULL,
  `ClassInSchool` text NOT NULL,
  `Number` int(11) NOT NULL,
  `SchoolClub` text NOT NULL,
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `GoSchool` (
  `uid` VARCHAR(100) NOT NULL,
  `name` varchar(10) NOT NULL,
  `email` text NOT NULL,
  `photoURL` text NOT NULL,
  `GradeInSchool` text NOT NULL,
  `ClassInSchool` text NOT NULL,
  `Number` int(11) NOT NULL,
  `lateness` boolean DEFAULT FALSE,
  `latenessTime` VARCHAR(1) DEFAULT 0,
  `EarlyBack` boolean DEFAULT FALSE,
  `EarlyBackTime` VARCHAR(1) DEFAULT 0,
  `Date` date NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE USER IF NOT EXISTS 'MiraiCore'@'%' IDENTIFIED BY 'KT34i5kirQpV';
GRANT ALL ON `ITmiraiApp`.* TO 'MiraiCore'@'%';