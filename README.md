# aws messaging

## aws-msg library

This library is for integrating Amazon Web Services (AWS) Simple Notification Service (SNS) with Simple Queue Service (SQS) to build a messaging infrastructure. Description of `aws-msg` library:

**Components and Their Responsibilities:**

1. **Exchange Struct and Methods:**
   - The `Exchange` struct represents an AWS SNS topic and encapsulates its interactions.
   - It provides methods to manage the SNS topic, subscriptions, and message publishing.
   - Responsibilities:
      - Create and manage an SNS topic.
      - Subscribe an SQS queue to the SNS topic.
      - Set subscription attributes, such as filter policies.
      - Publish messages to the SNS topic.

2. **SqsMediator Struct and Methods:**
   - The `SqsMediator` struct represents a system for managing SQS queues and their interactions with SNS.
   - It provides methods for managing SQS queues, subscribing them to SNS topics, and receiving messages from SQS queues.
   - Responsibilities:
      - Create and manage SQS queues, including their attributes.
      - Subscribe SQS queues to SNS topics.
      - Poll SQS queues for incoming messages and process them.
      - Remove messages from SQS queues.

**Key Features and Functionalities:**

- **Exchange Management:**
   - The `aws-msg` library allows the creation and management of SNS topics (represented by the `Exchange` struct).
   - Topics can be customized with a suffix for unique identification.

- **Subscription Management:**
   - The `aws-msg` library enables the subscription of SQS queues to SNS topics.
   - Subscriptions can be configured with attributes, including filter policies.

- **Message Handling:**
   - The `aws-msg` library supports publishing messages to SNS topics.
   - Messages are constructed with attributes, including an event type.
   - Subscribed SQS queues can poll for incoming messages and process them.
   - Messages can be removed from SQS queues after processing.

**Use Cases:**

1. **Messaging System:** The `aws-msg` library is suitable for building messaging systems where various components need to exchange messages through SNS and SQS.

2. **Event Notification:** It can be used to send event notifications from one component to another, allowing decoupled communication.

3. **Message Processing:** The `aws-msg` library can be employed to process incoming messages asynchronously, ensuring fault tolerance and scalability.

**Advantages:**

- **Simplified Integration:** The `aws-msg` library abstracts the complexities of AWS SNS and SQS interactions, providing a structured and easy-to-use API.

- **Flexibility:** It allows customization of topics, subscriptions, and message attributes to meet specific requirements.

- **Decoupled Architecture:** SNS and SQS enable decoupled communication between components, improving system reliability.

**Potential Improvements:**

- **Error Handling:** Implement comprehensive error handling throughout the `aws-msg` library to handle AWS service errors gracefully.

- **Logging:** Incorporate a logging mechanism to capture and log events, errors, and message processing details.

- **Security:** Enhance security practices by managing AWS credentials securely, implementing access controls, and using encryption.

- **Testing:** Develop unit tests and integration tests to ensure the reliability of the `aws-msg` library.

- **Documentation:** Provide clear documentation and usage examples to assist developers in using the library effectively.

**Comparison with RabbitMQ:**

- RabbitMQ is a message broker that provides robust queuing and message routing capabilities, while the provided `aws-msg` library leverages AWS SNS and SQS services.
- RabbitMQ offers more advanced features like message persistence, message routing rules, and fanout exchanges, while AWS SNS/SQS offers easy scalability and managed services.
- AWS SNS/SQS is fully managed, while RabbitMQ requires self-hosting or managed service setup.

**Limitations:**

- AWS SNS/SQS is a cloud-based solution with associated costs, whereas RabbitMQ can be deployed on-premises or in various cloud environments.

- The `aws-msg` library is tightly integrated with AWS services, making it less portable to other cloud providers.

- AWS SNS/SQS is eventually consistent, which may not be suitable for real-time applications that require strict ordering and consistency guarantees.

- While the `aws-msg` library provides a useful abstraction, it may not cover all advanced messaging scenarios that RabbitMQ can handle.

In summary, the provided `aws-msg` library simplifies AWS SNS and SQS integration for building messaging infrastructure, offering advantages in terms of ease of use and scalability. However, it's important to consider the specific requirements and use cases to determine whether this `aws-msg` library or an alternative like RabbitMQ is the best fit for a particular application.

----
## **Key Components:**
    
- **Exchange:** The `Exchange` struct represents an SNS topic. It serves as a central hub for publishing events to multiple subscribers (SQS queues). The `Exchange` struct provides methods for publishing messages to the topic and managing subscriptions.

- **SqsMediator:** The `SqsMediator` struct facilitates the creation and management of SQS queues and their subscriptions to SNS topics. It acts as a coordinator between the Exchange and individual subscribers.

- **Subscriber Interface:** The `Subscriber` interface defines the contract for creating custom message handlers. Subscribers must implement two methods: `GetSettings()` to configure the subscription and `Handle(msg *chassis.Message)` to process incoming messages.

- **Message:** The `Message` struct wraps an SQS message, providing convenience methods for acknowledging message receipt and removing it from the queue.

- **SubscriptionSettings:** This struct encapsulates subscription-related settings, including the SNS topic, SQS queue, Dead Letter Queue (DLQ) configuration, and optional subscription attributes like Filter Policies.

## **Sample Implementations:**

- **Creating an Exchange:**
  To create an Exchange (SNS topic), you typically use the `NewExchange` function. Here's an example:

  ```go
  session, _ := session.NewSession(&aws.Config{
      Region: aws.String("us-east-1"),
  })

  settings := chassis.SubscriptionSettings{
      Topic: "MyTopic",
  }

  exchange, err := chassis.NewExchange(session, "MySuffix", settings)
  if err != nil {
      log.Fatal(err)
  }
  ```

- **Creating a Subscriber:**
  Implement the `Subscriber` interface to create custom message handlers. Here's an example:

  ```go
  type MySubscriber struct {
      name string
  }

  func (s MySubscriber) GetSettings() chassis.SubscriptionSettings {
      return chassis.SubscriptionSettings{
          Topic: "MyTopic",
          Queue: s.name,
      }
  }

  func (s MySubscriber) Handle(msg *chassis.Message) error {
      log.Infof("Received message: %s", msg.Body.Message)
      return msg.Ack(nil)
  }
  ```

- **Creating a SqsMediator:**
  Use the `SqsMediator` to manage subscriptions and handle incoming messages. Here's an example:

  ```go
  session, _ := session.NewSession(&aws.Config{
      Region: aws.String("us-east-1"),
  })

  mediator := chassis.NewSqsMediator(session, "MyService", "MySuffix", 5)
  ```

- **Subscribing to an Exchange:**
  Subscribe a custom subscriber to an Exchange:

  ```go
  subscriber := MySubscriber{name: "MyQueue"}
  err := subscriber.Subscribe(subscriber.GetSettings())
  if err != nil {
      log.Fatal(err)
  }
  ```

- **Receiving Messages:**
  Use the `Receive` method to retrieve and process messages:

  ```go
  settings := chassis.SubscriptionSettings{
      Topic: "MyTopic",
      Queue: "MyQueue",
  }

  messages, err := subscriber.Receive(settings)
  if err != nil {
      log.Fatal(err)
  }

  for _, message := range messages {
      // Process each message asynchronously
      go func(msg *chassis.Message) {
          if err := subscriber.Handle(msg); err != nil {
              log.Errorf("Error processing message: %v", err)
          }
      }(message)
  }
  ```
  
 ### Sample implementation available
- [publisher](./publisher)
- [subscriber](./subscriber)


## **Usage in AWS Ecosystem:**
- Deploy your subscribers as AWS Lambda functions, EC2 instances, or other AWS services.
- Configure SNS topics to publish events to subscribers based on your application's event-driven architecture.
- Ensure proper IAM permissions to allow SNS and SQS access for your application components.

## **Benefits:**
- Simplifies event-driven communication within AWS.
- Decouples components of your AWS application.
- Supports message filtering with Filter Policies.
- Handles failures gracefully with Dead Letter Queues.

By following this pattern and using the `aws-msg` library, you can build scalable and reliable event-driven systems within the AWS ecosystem, enabling seamless communication between components while maintaining flexibility and robustness.