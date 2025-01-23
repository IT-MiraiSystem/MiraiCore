CREATE DATABASE IF NOT EXISTS `ITmiraiApp`;

USE `ITmiraiApp`;

CREATE TABLE IF NOT EXISTS `ChangeOfClass` (
  `Date` date NOT NULL,
  `DayOfTheWeek` varchar(20) NOT NULL,
  `ClassID` varchar(5) NOT NULL,
  `LeasonNumber` int(10) NOT NULL,
  `Lesson` VARCHAR(20) NOT NULL,
  `Room` varchar(100) NOT NULL,
  `Teacher` varchar(10) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `ClassTimetable`(
  `ClassID` varchar(5) NOT NULL,
  `DayOfTheWeek` varchar(20) NOT NULL,
  `Lesson` VARCHAR(20) NOT NULL,
  `LeasonNumber` int(10) NOT NULL,
  `Room` varchar(100) NOT NULL,
  `Teacher` varchar(10) NOT NULL,
  `StartTime` time NOT NULL,
  `EndTime` time NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `Users` (
  `uid` VARCHAR(100) NOT NULL,
  `name` VARCHAR(10) NOT NULL,
  `email` TEXT NOT NULL,
  `photoURL` TEXT NOT NULL,
  `GradeInSchool` TEXT NOT NULL,
  `ClassInSchool` TEXT NOT NULL,
  `Number` INT(11) NOT NULL,
  `SchoolClub` TEXT,
  `location` TEXT,
  `Permission` INT(1) DEFAULT 0,
  `Subject` JSON,
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `Issue` (
  `ClassID` VARCHAR(5) NOT NULL,
  `Issue` TEXT NOT NULL,
  `Lesson` TEXT NOT NULL,
  `Period` TEXT NOT NULL,
  `Submitter` JSON
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `Event`(
  `ClassID` varchar(5) NOT NULL,
  `Event` TEXT NOT NULL,
  `Date` date NOT NULL
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
  `Date` date NOT NULL,
  `CommuteTime` date NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `attendance`(
  `ClassID` varchar(5) NOT NULL,
  `Leasson` VARCHAR(20) NOT NULL,
  `Date` date NOT NULL,
  `LeasonNumber` int(10) NOT NULL,
  `Attendance` JSON
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE USER IF NOT EXISTS 'MiraiCore'@'%' IDENTIFIED BY 'KT34i5kirQpV';
GRANT ALL ON `ITmiraiApp`.* TO 'MiraiCore'@'%';