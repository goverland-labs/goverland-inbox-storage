# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2024-11-01

### Added
- Added github actions
- Added README.md
- Added CONTRIBUTING.md
- Added LICENSE

## [0.4.10] - 2024-10-08

### Changed
- Enable by default ArchiveProposalAfterVote setting

## [0.4.9] - 2024-10-02

### Added
- Added user delegation
- Added delegate allowed daos

## [0.4.8] - 2024-09-13

### Added
- Added github actions for running tests, linter and building docker image

## [0.4.7] - 2024-07-26

### Fixed
- Disable link AI summarization

## [0.4.6] - 2024-07-26

### Fixed
- Update prompt for the AI client part 3

## [0.4.5] - 2024-07-26

### Fixed
- Update AI summary model

## [0.4.4] - 2024-07-26

### Added
- Logging AI summary requests

## [0.4.3] - 2024-07-26

### Fixed
- Update prompt for the AI client part 2

## [0.4.2] - 2024-07-26

### Fixed
- Update prompt for the AI client

## [0.4.1] - 2024-07-23

### Fixed
- Getting recent dao by raw query instead of filters

## [0.4.0] - 2024-07-22

### Added
- Implement app versions endpoint

## [0.3.0] - 2024-07-05

### Added
- Implement AI summary endpoint

## [0.2.1] - 2024-07-05

### Changed
- Default autoarchive days from 7 to 1

## [0.2.0] - 2024-07-05

### Added
- Implement feed settings server methods

## [0.1.0] - 2024-07-01

### Added
- Implement push settings server methods


## [0.0.34] - 2024-06-14

### Fix
- Comparing APP versions in achievements

## [0.0.33] - 2024-06-13

### Changed
- Separate storing tokens by device

### Added
- Endpoint for getting list of tokens

## [0.0.32] - 2024-05-13

### Fixed
- Hide achievements for guests

## [0.0.31] - 2024-05-03

### Fixed
- Prefill subscribers list on application bootstrap

## [0.0.30] - 2024-04-29

### Fixed
- Getting voting stats from cache

## [0.0.29] - 2024-04-28

### Fixed
- Fixed getting user by address method

## [0.0.28] - 2024-04-26

### Changed
- Update status code if user profile was not found

## [0.0.27] - 2024-04-19

### Changed
- Use recommendations from core storage instead of prepared data

## [0.0.26] - 2024-04-16

### Changed
- Add micro optimization for getting votes for user
- Achievements sorting

## [0.0.25] - 2024-04-11

### Added
- Achievements implementation

## [0.0.24] - 2024-03-26

### Changed
- Actualize ens proto library

## [0.0.23] - 2024-03-25

### Added
- Add user by address

## [0.0.22] - 2024-03-22

### Added
- Integrate with ZerionAPI for dao recommendations

## [0.0.21] - 2024-03-15

### Added
- Achievements schema

### Added
- Added featured proposals implementation

## [0.0.20] - 2024-03-06

### Added
- Nats metrics

## [0.0.19] - 2024-02-28

### Added
- Added can vote calculation on user sign in

## [0.0.18] - 2024-02-14

### Added
- Added migration for user session: set last activity at not null 

## [0.0.17] - 2024-02-05

### Added
- Added me can vote feature implementation

## [0.0.16] - 2024-02-05

### Added
- Calculating intervals for sending push notifications

## [0.0.15] - 2024-01-31

### Added
- Tracking user activity endpoint
- Added user nonce usage for replay protection

## [0.0.14] - 2023-12-29

### Added
- Added ens name to user profile

## [0.0.13] - 2023-12-20

### Added
- Added user sessions, roles
- Suppor new user api

## [0.0.12] - 2023-11-07

### Changed
- Subscribe request also feed subscription

## [0.0.11] - 2023-11-03

### Added
- Improve logging

## [0.0.10] - 2023-11-01

### Added
- Implement recently viewed protocol

## [0.0.9] - 2023-07-27

### Added
- Added method for getting push token

## [0.0.8] - 2023-07-15

### Added
- Added method for fetching subscribers by DAO

## [0.0.7] - 2023-07-15

### Fixed
- Updated platform-events dependency to v0.0.20

## [0.0.6] - 2023-07-14

### Fixed
- Updated core-web-sdk dependency to v0.0.7

### Changed
- Changed sql schema of the database

## [0.0.5] - 2023-07-12

### Fixed
- Updated platform-events dependency to v0.0.13

## [0.0.4] - 2023-07-11

### Fixed
- Updated platform-events dependency to v0.0.11

## [0.0.3] - 2023-07-11

### Fixed
- Fixed Dockerfile
- Fixed linter warnings
- Fixed .gitignore file

### Added
- User settings

## [0.0.2] - 2023-07-03

### Added
- User subscriptions

## [0.0.1] - 2023-06-26

### Added
- Skeleton app
- Users with subscriptions 
