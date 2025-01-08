CREATE DATABASE "ITmiraiApp";

USE "ITmiraiApp";

CREATE TABLE "ChangeOfClass" (
  "date" date NOT NULL,
  "class" varchar(5) NOT NULL,
  "time" varchar(10) NOT NULL,
  "After" text NOT NULL,
  "label" text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE "PhoneOAuth" (
  "DeviceName" varchar(10) NOT NULL,
  "Pass" text NOT NULL,
  PRIMARY KEY ("DeviceName")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE "Users" (
  "uid" int(11) NOT NULL,
  "name" varchar(10) NOT NULL,
  "photoURL" text NOT NULL,
  "ClassInSchool" int(11) NOT NULL,
  "GradeInSchool" int(11) NOT NULL,
  "email" text NOT NULL,
  "SchoolClub" text NOT NULL,
  "FriendCode" int(11) NOT NULL,
  "FriendList" longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY ("uid")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE USER 'MiraiCore'@'%' identified by 'KT34i5kirQpV';
GRANT ALL on ITmiraiApp.* to 'MiraiCore'@'%';