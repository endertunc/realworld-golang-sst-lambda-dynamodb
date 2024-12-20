![RealWorld Example App](https://raw.githubusercontent.com/gothinkster/realworld-starter-kit/refs/heads/master/logo.png)

### Golang + SST + Lambda + DynamoDB + OpenSearch codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.

This codebase was created to demonstrate a backend application built with **Golang + SST + Lambda + DynamoDB + OpenSearch** including CRUD operations, authentication, routing, pagination, and more.

For more information on how this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.

## Architecture

### Overview
![image Architectural_Overview](./docs/architectural_overview.png)

#### API Flow
1. **API Gateway**
   - Entry point for all API requests

2. **Lambda Functions**
   - Separate handler for each API endpoint which implements business logic for specific operations

3. **DynamoDB**
   - Primary database for storing user, articles and comments

4. **OpenSearch Service**
   - Used for global queries such as most recent articles and list tags operations. It's utilizing DynamoDB zero-ETL integration with Amazon OpenSearch Service. 

#### Event Flow
1. **OpenSearch Ingestion Pipeline**
   - Processes article updates from DynamoDB Streams and indexes them in OpenSearch

2. **User Feed System**
   - DynamoDB Streams capture article changes
   - Feed Handler Lambda processes these changes and updates user feeds in real-time in Feed Table


### Local Development

![image Architectural_Overview](./docs/local_development_networking.png)

During local development, it's required that lambda functions can connect to the internet 
so that SST Live Lambda can communicate with your local machine. 
Therefore, in local development setup, we also deploy Internet Gateway (not required in production) to give Lambda functions access to the internet.
See https://docs.sst.dev/live-lambda-development for more details about SST Live Lambda.

### Production Networking

> Production Networking setup is NOT implemented yet. 
> I just need to deploy resources to different subnets and configure security groups accordingly depending on the stage that's all.
> I still wanted to share how the production networking diagram would look like.

![image Production Networking](./docs/production_networking.png)

## DynamoDB & OpenSearch

Dynamodb is used as the primary database for storing user, articles, and comments, etc. 
It covers the majority of the application access patterns.
However, for global queries such as _most recent articles_ and _list all tags_ operations,
I decided to use OpenSearch Service. 
One of the reasons is that I also wanted to experiment with Dynamodb Streams and OpenSearch Service Zero ETL Integration,
and I think it worked beautifully.


## DynamoDB Access Patterns

### User Table

#### Table Structure
```
Table Name: user

Primary Records:
- pk (STRING, Partition Key)  # Format: UUID
- email (STRING)             # User's email address
- username (STRING)          # User's username
- hashedPassword (STRING)    # Bcrypt hashed password
- bio (STRING, Optional)     # User's bio
- image (STRING, Optional)   # User's profile image URL
- createdAt (NUMBER)         # Unix timestamp
- updatedAt (NUMBER)         # Unix timestamp

Uniqueness Records:
- pk (STRING, Partition Key) # Format: "email#[email]" or "username#[username]"
                            # These records ensure email and username uniqueness

Global Secondary Indexes:
1. user_email_gsi
   - Partition Key: email
   - Projection: ALL

2. user_username_gsi
   - Partition Key: username
   - Projection: ALL
```

#### Access Patterns

| Index Used | Operation | Key Condition | Implementation Details |
|------------|-----------|---------------|----------------------|
| Primary Table (UUID) | Get User by ID | pk = [UUID] | - GetItem operation<br>- Strongly consistent read |
| | Get Multiple Users | Multiple pks | - BatchGetItem operation<br>- Used for following/follower lists |
| Primary Table (email#) | Create User | pk = "email#[email]" | - Part of TransactWriteItems<br>- Condition: attribute_not_exists(pk) |
| | Update User Email | pk = "email#[email]" | - Part of TransactWriteItems<br>- Delete old + Put new |
| Primary Table (username#) | Create User | pk = "username#[username]" | - Part of TransactWriteItems<br>- Condition: attribute_not_exists(pk) |
| | Update Username | pk = "username#[username]" | - Part of TransactWriteItems<br>- Delete old + Put new |
| user_email_gsi | Get User by Email | email = :email | - Query operation<br>- Returns all user attributes |
| user_username_gsi | Get User by Username | username = :username | - Query operation<br>- Returns all user attributes |

#### Design Considerations
   - Email uniqueness enforced by "email#[email]" records
   - Username uniqueness enforced by "username#[username]" records
   - TransactWriteItems ensures atomic operations for maintaining consistency

### Article Table

#### Table Structure
```
Table Name: article

Primary Records:
- pk (STRING, Partition Key) # Format: UUID
- title (STRING)             # Article title
- slug (STRING)              # URL-friendly version of title
- description (STRING)       # Article description
- body (STRING)              # Article content
- tagList (STRING[])         # Array of tags
- favoritesCount (NUMBER)    # Number of favorites
- authorId (STRING)          # UUID of the author
- createdAt (NUMBER)         # Unix timestamp
- updatedAt (NUMBER)         # Unix timestamp

Uniqueness Records:
- pk (STRING, Partition Key) # Format: "slug#[slug]"
                             # These records ensure slug uniqueness

Global Secondary Indexes:
1. article_slug_gsi
   - Partition Key: slug
   - Projection: ALL

2. article_author_gsi
   - Partition Key: authorId
   - Sort Key: createdAt
   - Projection: ALL
```

#### Access Patterns

| Index Used | Operation | Key Condition | Implementation Details |
|------------|-----------|---------------|----------------------|
| Primary Table (UUID) | Get Article by ID | pk = [UUID] | - GetItem operation<br>- Strongly consistent read |
| | Get Multiple Articles | Multiple pks | - BatchGetItem operation<br>- Used for feed and favorites |
| | Update Favorite Count | pk = [UUID] | - UpdateItem operation<br>- Atomic increment/decrement<br>- Part of favorite/unfavorite transaction |
| Primary Table (slug#) | Create Article | pk = "slug#[slug]" | - Part of TransactWriteItems<br>- Condition: attribute_not_exists(pk) |
| | Update Article Slug | pk = "slug#[slug]" | - Part of TransactWriteItems<br>- Delete old + Put new |
| article_slug_gsi | Get Article by Slug | slug = :slug | - Query operation<br>- Returns all article attributes |
| article_author_gsi | Get Articles by Author | authorId = :authorId | - Query operation<br>- Sort by createdAt<br>- Supports pagination |

#### Design Considerations
   - Slug uniqueness enforced by "slug#[slug]" records in the primary table
   - TransactWriteItems ensures atomic operations for maintaining consistency

### Comment Table

#### Table Structure
```
Table Name: comment

Attributes:
- commentId (STRING, Partition Key)  # UUID of the comment
- articleId (STRING, Sort Key)       # UUID of the article
- authorId (STRING)                  # UUID of the comment author
- body (STRING)                      # Comment content
- createdAt (NUMBER)                 # Unix timestamp
- updatedAt (NUMBER)                 # Unix timestamp

Global Secondary Indexes:
1. comment_article_gsi
   - Partition Key: articleId
   - Sort Key: createdAt
   - Projection: ALL
```

#### Access Patterns

| Index Used | Operation | Key Condition | Implementation Details |
|------------|-----------|---------------|----------------------|
| Primary Table | Create Comment | commentId + articleId | - PutItem operation<br>- Composite key ensures uniqueness |
| | Get Single Comment | commentId + articleId | - GetItem operation<br>- Strongly consistent read |
| | Delete Comment | commentId + articleId | - DeleteItem operation |
| comment_article_gsi | Get Comments by Article | articleId = :articleId | - Query operation<br>- Sort by createdAt<br>- Returns all comments |

#### Design Considerations
   - Each comment is directly linked to both its article and author
   - Article comments are partitioned by article via GSI and allow efficient retrieval of all comments for an article by creation date

### Favorite Table

#### Table Structure
```
Table Name: favorite

Attributes:
- userId (STRING, Partition Key)    # UUID of the user
- articleId (STRING, Sort Key)      # UUID of the article
- createdAt (NUMBER)                # Unix timestamp

Global Secondary Indexes:
1. favorite_user_id_created_at_gsi
   - Partition Key: userId
   - Sort Key: createdAt
   - Projection: ALL
```

#### Access Patterns

| Index Used | Operation | Key Condition | Implementation Details |
|------------|-----------|---------------|----------------------|
| Primary Table | Favorite Article | userId + articleId | - TransactWriteItems:<br>  1. Create favorite record<br>  2. Increment article favoritesCount |
| | Unfavorite Article | userId + articleId | - TransactWriteItems:<br>  1. Delete favorite record<br>  2. Decrement article favoritesCount |
| | Check Favorites | Multiple (userId + articleId) | - BatchGetItem operation |
| favorite_user_id_created_at_gsi | Get User Favorites | userId = :userId | - Query operation<br>- Sort by createdAt<br>- Supports pagination |

#### Design Considerations
   - Composite key in Favorite table ensures one favorite per user-article pair
   - User's favorite articles are partitioned by user via GIS and allow efficient retrieval of all favorite articles for a user by creation date

### Follower Table

#### Table Structure
> _There is no use case, but we should definitely store createdAt in this table to as general practice._
```
Table Name: follower

Attributes:
- follower (STRING, Partition Key)  # UUID of the user who is following
- followee (STRING, Sort Key)       # UUID of the user being followed
```

#### Access Patterns

| Index Used | Operation | Key Condition | Implementation Details |
|------------|-----------|---------------|----------------------|
| Primary Table | Follow User | follower + followee | - PutItem operation<br>- Creates follower relationship |
| | Unfollow User | follower + followee | - DeleteItem operation<br>- Removes follower relationship |
| | Check Following | follower + followee | - Query operation<br>- Uses SELECT COUNT<br>- Returns true if relationship exists |
| | Get Followees | Multiple (follower + followee) | - BatchGetItem operation<br>- Bulk check of following relationships |

#### Design Considerations
   - Composite key (follower + followee) ensures the unique following relationships

## Project Structure

```
.
├── cmd/                                  
│   └── functions/                        # API endpoint per Lambda function and event handlers
│       ├── add_comment/                  
│       ├── delete_article/               
│       ├── delete_comment/               
│       ├── favorite_article/             
│       ├── follow_user/                  
│       ├── get_article/                  
│       ├── get_article_comments/         
│       ├── get_current_user/             
│       ├── get_tags/                     
│       ├── get_user_feed/                
│       ├── get_user_profile/             
│       ├── list_articles/                
│       ├── login_user/                   
│       ├── post_article/                 
│       ├── register_user/                
│       ├── swagger/                      
│       ├── unfavorite_article/           
│       ├── unfollow_user/                
│       ├── update_article/               
│       ├── update_user/                  
│       └── user_feed/                    
├── internal/                             # Internal packages
│   ├── api/                              # API layer
│   │   ├── openapi/                      # OpenAPI/Swagger specifications
│   │   ├── article_api.go                
│   │   ├── comment_api.go                
│   │   ├── feed_api.go                   
│   │   ├── profile_api.go                
│   │   ├── user_api.go                   
│   │   ├── middleware.go                 # HTTP middleware (auth, logging)
│   │   ├── pagination.go                 # Pagination utilities
│   │   ├── request_helpers.go            # Request parsing and validation
│   │   └── response_helpers.go           # Response utilities
│   ├── database/                         # DynamoDB and OpenSearch clients
│   │   ├── dynamodb.go                   
│   │   └── opensearch.go                 
│   ├── errutil/                          # Error handling types and utilities
│   │   └── error.go                      
│   ├── repository/                       # Data access layer
│   │   ├── article_repository.go         
│   │   ├── comment_repository.go         
│   │   ├── feed_repository.go            
│   │   ├── follower_repository.go        
│   │   ├── user_repository.go            
│   │   └── mocks/                        # Repository mocks for testing
│   ├── security/                         # Security utilities
│   │   ├── auth.go                       # Authentication helpers for net/http
│   │   └── jwt.go                        # JWT token handling
│   ├── service/                          # Business logic layer
│   │   ├── article_service.go            
│   │   ├── article_list_service.go       
│   │   ├── comment_service.go            
│   │   ├── feed_service.go               
│   │   ├── profile_service.go            
│   │   ├── user_service.go               
│   │   └── mocks/                        # Service mocks for testing
│   └── test/                             # Testing utilities and entity helpers to support E2E tests
│       ├── article_entity_helper.go      
│       ├── auth_test_suite.go            
│       ├── comment_entity_helper.go      
│       ├── helpers.go                    
│       └── user_entity_helper.go         
├── stacks/                               # SST Infrastructure
│   ├── APIStack.ts                       # API Gateway and Lambda config
│   ├── DynamoDBStack.ts                  # DynamoDB tables and indexes
│   ├── OpenSearchStack.ts                # OpenSearch configuration
│   └── VPCStack.ts                       # VPC and network config
├── tools/                                # Development tools
│   └── jwt/                              # JWT key generation for local development
│   └── openapi/                          # OpenAPI specs generation
├── go.mod                                
├── Makefile                              # Build and development commands
├── package.json                          
├── sst.config.ts                         # SST configuration
└── tsconfig.json                         
```

### Internal Package Details

#### API Layer (`internal/api/`)
- OpenAPI/Swagger specifications for API documentation
- Request/response handling and validation
- Middleware for authentication, logging, and error handling
- Endpoint implementations for articles, comments, users, and profiles
- Pagination utilities for list endpoints

#### Database Layer (`internal/database/`)
- DynamoDB and OpenSearch client used by repository layer

#### Repository Layer (`internal/repository/`)
- Implementation of data access patterns for domain models
- DynamoDB and OpenSearch database models and mapping between database and domain models
- Mock implementations for testing

#### Service Layer (`internal/service/`)
- Core business logic implementation
- User profile and authentication logic
- Article&Comment management and operations
- Feed generation and filtering

#### Security Layer (`internal/security/`)
- JWT token generation and validation
- Authentication utilities
- Password hashing and verification

#### Testing Utilities (`internal/test/`)
- Test helpers and fixtures
- Common test suite setup

Each Lambda function in the `cmd/functions` directory contains the following files:
- `[function_name].go` - Main handler code for API endpoint or event handler
- `[function_name]_test.go` - E2E tests for the Lambda function

The `internal` packages contain the following subdirectories:
- `domain` - Business entities and interfaces
- `repository` - Data access implementation
- `service` - Business logic implementation
- `api` - HTTP request handling
