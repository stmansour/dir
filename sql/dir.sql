DROP DATABASE IF EXISTS accord;
CREATE DATABASE accord;
USE accord;

CREATE TABLE people (
    LastName VARCHAR(25),
    FirstName VARCHAR(25),
    MiddleName VARCHAR(25),
    Salutation VARCHAR(10),
    Class VARCHAR(25),
    Status VARCHAR(25),
    Employer VARCHAR(25),
    Title VARCHAR(50),
    Department VARCHAR(25),
    PositionControlNumber VARCHAR(10),
    Email VARCHAR(35),
    OfficePhone VARCHAR(25),
    OfficeFax VARCHAR(25),
    CellPhone VARCHAR(25),
    PrimaryEmail VARCHAR(35),
    SecondaryEmail VARCHAR(35),
    Hire varchar(20),
    Termination varchar(20),
    EligibleForRehire char(10),
    ReportsTo VARCHAR(25),
    LastReview varchar(20),
    NextReview varchar(20),
    Birthdate VARCHAR(25),
    HomeStreetAddress VARCHAR(35),
    HomeStreetAddress2 VARCHAR(25),
    HomeCity VARCHAR(25),
    HomeState CHAR(2),
    HomePostalCode varchar(10),
    HomeCountry VARCHAR(25),
    CompensationType VARCHAR(25),
    Deductions VARCHAR(100),
    HealthInsuranceAccepted Char(2),
    DentalInsuranceAccepted Char(2),
    Accepted401K CHAR(2)
);

source initpeople.sql

ALTER TABLE people add column uid MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY;
SELECT uid, lastname, firstname from people;
CREATE TABLE deductions (
    uid MEDIUMINT NOT NULL,
    deduction INT NOT NULL
    );

CREATE TABLE jobtitles (
    jobcode MEDIUMINT NOT NULL AUTO_INCREMENT,
    Title VARCHAR(40),
    Department VARCHAR(25),
    DeptCode MEDIUMINT,
    PRIMARY KEY (jobcode)
);

CREATE TABLE departments (
    deptcode MEDIUMINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(25),
    PRIMARY KEY (deptcode)
);

CREATE TABLE companies (
    cocode MEDIUMINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(25),
    address VARCHAR(35),
    address2 VARCHAR(25),
    city VARCHAR(25),
    state CHAR(2),
    postalcode varchar(10),
    designation char(3),
    country VARCHAR(25),
    Phone VARCHAR(25),
    Fax VARCHAR(25),
    Email VARCHAR(35),
    PRIMARY KEY (cocode)
);

CREATE TABLE compensation (
    uid MEDIUMINT NOT NULL,
    type MEDIUMINT NOT NULL
);

UPDATE people SET EligibleForRehire='Yes' WHERE EligibleForRehire="";

CREATE TABLE DeductionList (
    dcode MEDIUMINT NOT NULL,
    name VARCHAR(25)
);
