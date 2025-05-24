# Anonymous chat bot

Telegram bot that facilitates anonymous conversations between users. 
It offers features like connecting to chat partners, ending chats and checking chat status
all while maintaining user privacy.

## Commands
- `/start`: Get started with the bot and see the welcome message.
- `/connect` - Find someone to chat with.
- `/stop` - End the current chat session.
- `/help`: Get a quick guide on how to use the bot.
- `/status` - Check your chat connection status.
- `/next` - Find a new chat partner (Note: Currently a placeholder, will inform the user that the feature is not implemented).

## Typical User Flows

This section describes common ways users interact with the bot.

### 1. Starting a New Chat (Successful Connection)

*   **User A sends `/start` or `/connect`:**
    *   If User A sends `/start`, they receive a welcome message with a "Connect" button. Clicking it is equivalent to sending `/connect`.
    *   The bot checks if anyone is in the `WaitingQueue`.
*   **Scenario 1: Someone (User B) is waiting.**
    *   User A is immediately connected to User B.
    *   User A receives: "You are now connected to a random stranger!"
    *   User B (who was waiting) receives: "You are now connected to a random stranger!"
    *   Both users' statuses are updated to "chatting".
*   **Scenario 2: No one is waiting (User A is the first).**
    *   User A is added to the `WaitingQueue`.
    *   User A receives: "⏳ Looking for a partner for you. Please wait."
    *   User A's status is updated to "waiting".
*   **Later, User B sends `/connect`:**
    *   The bot sees User A in the `WaitingQueue`.
    *   User B is connected to User A.
    *   User B receives: "You are now connected to a random stranger!"
    *   User A (who was waiting) receives: "You are now connected to a random stranger!"
    *   Both users' statuses are updated to "chatting".

### 2. Sending and Receiving Messages

*   **User A is connected to User B.**
*   **User A sends a message (text, photo, sticker, etc.):** "Hello there!"
*   The bot receives User A's message, identifies User B as the partner, and forwards the message to User B.
*   **User B receives:** "Hello there!"
*   User B can reply, and the process repeats in reverse.

### 3. Ending a Chat

*   **User A is connected to User B.**
*   **User A sends `/stop`:**
    *   The bot updates User A's status to "idle" and clears their partner information.
    *   User A receives: "💬 Chat ended. Type /connect to find someone else."
    *   The bot updates User B's status to "idle" and clears their partner information.
    *   User B receives: "Stranger has left the chat. Type /connect to find someone else."

### 4. Trying to Connect (No One Available)

*   **User A sends `/connect`:**
*   The `WaitingQueue` is empty.
*   User A is added to the `WaitingQueue`.
*   User A receives: "⏳ Looking for a partner for you. Please wait."
*   User A's status is updated to "waiting". They will remain in this state until another user connects or they use `/stop`.

### 5. Using `/stop` While in the Waiting Queue

*   **User A has previously sent `/connect` and is waiting for a partner.**
*   User A's status is "waiting".
*   **User A sends `/stop`:**
    *   User A is removed from the `WaitingQueue`.
    *   User A's status is updated to "idle".
    *   User A receives: "You have been removed from the connection queue."

### 6. Checking Status (`/status`)

*   **Scenario 1: User is disconnected (has not used `/connect` or has used `/stop`).**
    *   User sends `/status`.
    *   User receives: "You are not connected to anyone and not in the waiting list."
*   **Scenario 2: User is in the waiting queue.**
    *   User sends `/status`.
    *   User receives: "You are currently in the waiting list to be connected."
*   **Scenario 3: User is connected and chatting.**
    *   User sends `/status`.
    *   User receives: "You are currently chatting with a stranger."

## Architecture

The bot is composed of several key Go files:

*   **`main.go`**: This is the entry point of the application. It initializes the Telegram bot, sets up the command handlers, and starts the main event loop to listen for and process user messages and commands. It orchestrates the interactions between the user, the Telegram API, and the bot's internal logic.
*   **`queue/queue.go`**: This file implements a thread-safe queue (`WaitingQueue`) using a linked list data structure. This queue is used to hold users who are waiting to be matched with a chat partner. It provides operations like `Enqueue` (add a user to the queue) and `Dequeue` (remove a user from the queue).
*   **`store/user_store.go`**: This file defines the `UserStore`, which is responsible for managing the state of users. It's an in-memory map that stores user IDs as keys and their `User` objects as values. Each `User` object contains information about the user's current chat status (e.g., idle, waiting, chatting) and their partner's ID if they are in a chat. It provides thread-safe methods to get, set, and update user information.

## Functionality

### User Matching Process

1.  A user initiates a connection by sending the `/connect` command.
2.  The bot checks the `WaitingQueue` (`queue/queue.go`).
3.  **If the `WaitingQueue` is not empty**: The bot dequeues the user who has been waiting the longest. This user and the initiating user are then connected. Their statuses in the `UserStore` (`store/user_store.go`) are updated to "chatting", and their respective partner IDs are stored. Both users receive a notification that they have been connected.
4.  **If the `WaitingQueue` is empty**: The initiating user is enqueued into the `WaitingQueue`, and their status in the `UserStore` is updated to "waiting". They receive a message informing them that they are waiting for a partner.

### Message Relaying

1.  A user sends a message to the bot.
2.  The bot checks the user's status in the `UserStore`.
3.  If the user is in a "chatting" state, the bot retrieves their partner's ID from the `UserStore`.
4.  The bot then forwards the message directly to the partner's Telegram chat ID.

### Chat Session Management

*   **Starting a Chat:** Chat sessions are automatically started as part of the "User Matching Process" described above when two users are successfully paired.
*   **Stopping a Chat:**
    1.  A user issues the `/stop` command.
    2.  The bot checks the user's status in the `UserStore`.
    3.  If the user was in the `WaitingQueue`, they are removed from it, and their status is updated to "idle". They are notified that they are no longer waiting.
    4.  If the user was in a "chatting" state:
        *   The bot retrieves the partner's ID.
        *   Both the initiating user's and their partner's status in the `UserStore` are updated to "idle", and their partner ID fields are cleared.
        *   Both users are sent a notification that their chat session has ended.

### Data Storage and Implications

The bot currently utilizes in-memory data structures for managing its state:

*   **`UserStore`**: A Go map (`map[int64]*User`) stores user information.
*   **`WaitingQueue`**: A linked list (implemented in `queue/queue.go`) stores users waiting for a chat partner.

**Implications**: Because these data structures are in-memory, all chat session data, user statuses, and the waiting queue will be **lost if the bot restarts** (e.g., due to a crash, deployment, or server reboot). Users will need to reconnect and will lose their previous chat partners. For persistent storage, a database solution would be required.

## Potential Issues and Limitations

This section highlights known issues, limitations, and areas for future improvement.

1.  **Lack of Persistent Storage**:
    *   As detailed in the "Data Storage and Implications" section, the bot relies entirely on in-memory data structures. This means that if the bot restarts for any reason (crash, update, server maintenance), all current user states, active chats, and users in the waiting queue are lost. Users would need to initiate new `/connect` requests. This is a significant limitation for a production environment.

2.  **Potential Bug in `queue.RemoveNode`**:
    *   The `queue.RemoveNode(chatId int64)` function in `queue/queue.go` has a bug. It correctly handles cases where the node to be removed is the head of the queue. However, if the node to be removed is the *tail* of the queue (and not also the head), the function will fail to remove it and incorrectly return an error stating the `chatId` was not found.
    *   **Reason**: The loop `for current.Next != nil` iterates while `current` has a subsequent node. If the `chatId` matches `current.Next.ChatId`, the node is removed. The loop terminates when `current.Next` is `nil`, meaning `current` is the last node (the tail). The `ChatId` of this tail node itself is never compared against the `chatId` to be removed within the loop.

3.  **Unused `userStore.FindMatch` Function**:
    *   The `store/user_store.go` file contains a public method `FindMatch(excludeChatId int64) (*User, bool)`. This function appears designed to find a waiting user in the `UserStore`.
    *   However, this function is **not currently used** by the matching logic in `main.go`. The user matching process in `HandleConnect` (within `main.go`) relies solely on the `waitingQueue.Dequeue()` mechanism to find a partner.
    *   This suggests that `userStore.FindMatch` might be leftover code from a previous implementation, a feature that was planned but not fully integrated, or potentially redundant.

4.  **` /next ` Command Not Implemented**:
    *   The `/next` command is registered and appears in the command list, but its handler currently only returns a "feature not implemented" message. This provides a command for users that does not yet offer any functionality.

## Future Enhancements

This section outlines potential features and improvements that could be implemented in the future to enhance the bot's functionality, reliability, and user experience.

### 1. Fix Bug in `queue.RemoveNode`

*   **Problem**: The `queue.RemoveNode` function in `queue/queue.go` has a bug where it fails to remove the tail element of the queue if that element is not also the head. A detailed explanation of the bug's cause is in the "Potential Issues and Limitations" section (Issue #2).
*   **Impact**:
    *   If a user is the last one in the `WaitingQueue` (i.e., they are the tail), and the `/stop` command is issued for this user, `queue.RemoveNode` will fail to remove them.
    *   This can lead to an inconsistent state where the bot believes the user is still in the queue, but the user believes they have stopped waiting. It may also cause issues if the user tries to `/connect` again.
*   **Proposed Solution**: Revise the logic in `queue.RemoveNode` to correctly handle the removal of the tail element. This involves adding a check after the loop to see if the `current` node (which would be the tail at that point) is the one to be removed.
*   **Benefits**:
    *   **Increased Reliability**: Ensures users can always be correctly removed from the waiting queue.
    *   **State Consistency**: Prevents discrepancies between the bot's understanding of the queue and the user's status.
    *   **Improved User Experience**: Avoids potential confusion or errors for users who are the last in the queue and try to stop waiting.

### 2. Implement `/next` Command Functionality

*   **Problem**: The `/next` command is currently registered with the bot and visible to users, but it is non-functional. It only returns a "feature not implemented" message (as noted in "Potential Issues and Limitations", Issue #4).
*   **Impact**: Users might try to use `/next` expecting to quickly switch partners, but find it doesn't work, leading to a slightly confusing experience.
*   **Proposed Solution**: Implement the full functionality for the `/next` command to allow users to seamlessly transition from their current chat to a new one.
*   **Benefits**:
    *   **Improved User Experience**: Provides a much quicker and more intuitive way for users to find a new chat partner compared to the current method of manually using `/stop` and then `/connect`.
    *   **Enhanced Usability**: Makes the bot feel more dynamic and responsive to user desires for changing chat partners.
*   **Potential Logic**:
    1.  When a user issues the `/next` command:
        *   **If the user is currently in a chat**:
            *   The bot should perform the same actions as `/stop` for the current chat:
                *   Notify the current partner that the user has left.
                *   Update both users' statuses in the `UserStore` to "idle" and clear their partner information.
            *   Immediately proceed to the next step.
        *   **If the user is waiting in the queue**:
            *   They should be removed from their current position in the queue (to avoid immediate re-connection to an undesired spot if they were at the front).
            *   Immediately proceed to the next step.
        *   **If the user is idle (neither chatting nor waiting)**:
            *   Proceed directly to the next step.
    2.  The bot should then attempt to connect the user to a new partner:
        *   This would essentially be the same logic as the `/connect` command:
            *   Check the `WaitingQueue` for an available user.
            *   If a user is found, connect them.
            *   If the `WaitingQueue` is empty, add the initiating user to the queue.
    3.  Provide appropriate feedback messages to the user throughout this process (e.g., "Ending current chat...", "Looking for a new partner...").

### 3. Address Unused `userStore.FindMatch` Function

*   **Problem**: The `store/user_store.go` module contains a public function `FindMatch(excludeChatId int64) (*User, bool)`. This function iterates through the `UserStore` to find a user whose `IsConnecting` status is true. However, as noted in "Potential Issues and Limitations" (Issue #3), this function is not currently utilized by the user matching logic in `main.go`, which instead relies on `waitingQueue.Dequeue()` (a FIFO queue).
*   **Impact**: The presence of unused code can lead to confusion for developers, increase the cognitive load when trying to understand the system, and potentially hide outdated logic.
*   **Proposed Actions**: A decision should be made regarding the future of `userStore.FindMatch`:
    1.  **Integrate or Adapt**:
        *   Evaluate if the `FindMatch` logic could be beneficial for an alternative or more flexible matching strategy. For example, if user preferences or other criteria were introduced in the future, searching the `UserStore` directly might be more suitable than a simple FIFO queue.
        *   If a valid use case is identified, `FindMatch` could be integrated into the matching process, potentially alongside or replacing the current queue mechanism depending on the desired behavior.
    2.  **Remove**:
        *   If `FindMatch` is confirmed to be a remnant of a previous design, or if the current `WaitingQueue` approach is deemed sufficient and simpler for the bot's objectives, then `FindMatch` should be removed from `store/user_store.go`.
*   **Benefits of Addressing**:
    *   **Improved Code Clarity**: The codebase becomes easier to understand without dead or redundant code.
    *   **Enhanced Maintainability**: Reducing unused code simplifies future maintenance and refactoring efforts.
    *   **Clearer Design Intent**: The matching logic becomes more explicit and less ambiguous.

### 4. Basic Anonymous User Preferences (e.g., Avoid Last Partner)

*   **Concept**: Introduce optional, basic user preferences that can enhance the chat experience while strictly maintaining anonymity.
*   **Specific Suggestion**: **Option to Avoid Immediate Reconnection with Last Chat Partner.**
    *   **Problem**: Occasionally, users might use `/stop` (or a future `/next` command) and then immediately `/connect` again, only to be re-paired with the person they just left if that person is also at the front of the queue. This can be awkward or undesirable.
    *   **Proposed Solution**:
        1.  When a chat ends (via `/stop` or `/next`), temporarily and anonymously store the ex-partner's ID in the `User` object (e.g., as `lastConnectedPartnerId`). This ID should only be stored for a short, predefined duration or until the user connects to someone else.
        2.  When the user tries to connect again, and the bot dequeues a potential partner, the system would check if this potential partner's ID matches the `lastConnectedPartnerId`.
        3.  If it matches, the bot could temporarily skip this match and try to dequeue the next person in the queue (if available). If no other suitable partner is found, the user might be placed back in the queue or, depending on design choices, informed that only their last partner is available.
*   **Anonymity Considerations (Crucial)**:
    *   The storage of the `lastConnectedPartnerId` must be strictly temporary. It should not create any permanent link or chat history between users.
    *   This "avoidance" should be short-lived (e.g., for the next connection attempt within a few minutes, or only if the user immediately tries to reconnect).
    *   No user should be able to see a "list" of users they've avoided or who have avoided them. The mechanism should be entirely internal to the matching logic.
*   **Benefits**:
    *   **Improved User Satisfaction**: Reduces potentially awkward or unwanted immediate reconnections, giving users more control over their experience.
*   **Trade-offs**:
    *   **Increased Complexity**: This adds complexity to the matching logic and user state management.
    *   Careful design is needed to balance the feature with anonymity and fairness in the queue.
*   **Future Scope**: This could be a first step towards more sophisticated (but still anonymous) preference settings if desired, though each would need careful consideration regarding anonymity.

### 5. Basic Rate Limiting

*   **Purpose**: To protect the bot from spam or abuse, particularly of the matching system through commands like `/connect` and a functional `/next`. Malicious or malfunctioning clients could otherwise repeatedly trigger matching attempts.
*   **Impact of Abuse**:
    *   **Service Degradation**: Excessive requests from a single user (or a few users) can consume server resources (CPU, memory, network bandwidth) and slow down the matching process for all other users.
    *   **Unnecessary Server Load**: Leads to higher operational costs and potential instability.
    *   **Queue Distortion**: Rapid, repeated `/connect` / `/stop` / `/next` sequences could unfairly manipulate the waiting queue.
*   **Suggested Strategies**:
    *   **Command-Specific Limits**: Implement a limit on the number of times a specific user can execute certain commands (primarily `/connect` and `/next`) within a defined time window. For example:
        *   Allow a maximum of X `/connect` (or `/next`) attempts per minute per user.
        *   A user exceeding this limit would receive a temporary cooldown message, e.g., "You are trying to connect too frequently. Please wait a moment and try again."
    *   **Message Rate Limits**: While Telegram's API handles some message rate limiting, the bot could also enforce its own limits on incoming messages if message-based abuse becomes an issue (though command abuse is the primary concern for this point).
*   **Implementation Details**:
    *   This would likely involve storing timestamps of relevant command usage per user (in-memory, with considerations for persistence if that feature is added).
    *   The specific limits (e.g., number of attempts, time window duration) should be **configurable** (e.g., via environment variables or a configuration file). This allows for tuning based on observed usage patterns and bot performance.
*   **Benefits**:
    *   **Improved Stability and Fairness**: Protects the bot's resources and ensures a more equitable experience for all users.
    *   **Abuse Prevention**: Deters simple forms of spam and automated abuse.

### More Advanced Feature Ideas

The following ideas are more complex to implement and require careful consideration of anonymity and system load, but could significantly enhance the bot's capabilities.

### 1. Interest-Based Matching (Anonymous)

*   **Concept**: Allow users to optionally and anonymously select a few general, predefined interests. This could help facilitate more engaging conversations by pairing users who share common ground.
    *   Users could be presented with a predefined list of broad interest categories (e.g., "Movies," "Gaming," "Music," "Books," "Sports," "Technology," "Travel," "Food").
    *   This selection would be optional. Users who don't select interests would be matched randomly as per the current system.
*   **Matching Logic**:
    *   **Prioritization**: The matching system could prioritize pairing users who share one or more selected interests.
    *   **Multiple Interests**: If users can select multiple interests (e.g., up to 2-3), the system could try to match on as many shared interests as possible.
    *   **Fallback Options**:
        *   If no immediate match with shared interests is found, the user could be asked if they prefer to wait longer for an interest-based match or connect with any available user.
        *   Alternatively, after a certain waiting period for an interest-match, the system could automatically broaden the search to any available user.
*   **Anonymity Considerations (Crucial)**:
    *   **Predefined, Limited List**: Interests *must* be selected from a fixed, general list provided by the bot. Free-text interest fields are not permissible as they could easily compromise anonymity.
    *   **Limited Selection**: Users should only be allowed to select a small number of interests (e.g., maximum 2-3) to prevent overly specific profiles.
    *   **Interests Not Revealed**: The system must **never** explicitly reveal a user's selected interests (or their partner's interests) to either party in the chat. The connection should just seem like a slightly more compatible random match.
    *   **Avoiding Small "Interest Silos"**: The list of interests and the matching logic need to be designed to prevent the creation of very small groups of users with unique combinations of interests. If an interest or combination is too rare, it could inadvertently lead to deanonymization or make it easier to guess who a user might be. The system should ensure a sufficiently large pool for each interest category.
*   **Benefits**:
    *   **More Engaging Conversations**: Users are more likely to have common topics to discuss, potentially leading to more meaningful and longer-lasting anonymous interactions.
    *   **Increased User Retention**: A higher chance of positive chat experiences could encourage users to use the bot more frequently.
*   **Complexity**:
    *   **User Profiles (Anonymous)**: Requires adding a way to temporarily store these anonymous interest selections for users who are actively seeking a chat. This data must be handled with the same care as other user session data regarding anonymity and volatility (or persistence, if implemented).
    *   **Matching Algorithm**: The matching logic becomes significantly more complex. It would need to manage separate queues or a more sophisticated weighting system for users based on interests, handle fallbacks, and manage the anonymity considerations mentioned above.
    *   **User Interface**: A simple UI (e.g., inline keyboards) would be needed for users to select/deselect interests.

### 2. Simple Anonymous Group Chats

*   **Concept**: Allow small groups of users (e.g., 3-4, configurable) to connect into a single, temporary, and anonymous chat session. This provides an alternative to the 1-on-1 chat format.
*   **User Experience**:
    *   A user could initiate a group chat request via a command like `/connect_group` or by selecting a "Group Chat" option from an interactive menu.
    *   They would then be added to a dedicated "group waiting queue."
    *   Once a sufficient number of users (e.g., a minimum of 3) are in this queue, the bot would automatically form a group chat, and all participants would be notified.
*   **Functionality**:
    *   **Message Relaying**: Messages sent by any member of the group would be relayed by the bot to all other members of that specific group.
    *   **In-Group Identification (Optional)**:
        *   For clarity, the bot could assign temporary, anonymous identifiers for users within the group chat context (e.g., "User 1," "User 2," "User 3"). These identifiers would be unique per group session.
        *   Alternatively, for a more chaotic or "purely anonymous" experience, messages could be forwarded without any sender identifier, though this might make conversations harder to follow. This choice would be a design decision.
    *   **Leaving a Group**: A `/stop` or `/leave_group` command would remove a user from their current group chat.
        *   The group might continue with the remaining members if a minimum number is still present (e.g., at least 2).
        *   If a user leaving causes the group to fall below a minimum threshold, the group could be automatically disbanded, and remaining members notified.
*   **Anonymity Considerations**:
    *   **No User Profiles**: As with 1-on-1 chats, no actual user profiles, usernames, or Telegram IDs would be shared among group members.
    *   **Random Composition**: Group composition would be based on the random order of users joining the group waiting queue.
    *   **Internal Identifiers**: If in-group identifiers are used (e.g., "User 1"), they must be temporary, session-specific, and not linkable to any permanent identity.
*   **Benefits**:
    *   Offers a different social dynamic compared to 1-on-1 interactions.
    *   Could be interesting for short, casual, multi-person anonymous discussions or games.
    *   May cater to users looking for a slightly broader anonymous interaction.
*   **Complexity**:
    *   **Group State Management**: The bot needs to manage the state of active groups, including which users belong to which group.
    *   **New Queueing/Matching Logic**: A separate queueing and matching mechanism for forming groups would be required.
    *   **Message Fan-Out**: The message relaying logic needs to handle fanning out messages to multiple users in a group efficiently.
    *   **Dynamic Group Membership**: Handling users leaving (and potentially joining, though initial implementation might fix groups once formed) adds complexity to maintaining group integrity and notifying members.

### 3. Timed Chat Option

*   **Concept**: Provide an option for users, once connected in a 1-on-1 chat, to mutually agree to a timed chat session. This would add a predefined duration to their interaction (e.g., 10, 15, or 30 minutes).
*   **User Experience**:
    *   **Initiation**: Shortly after a successful connection, the bot could privately ask each user (e.g., via an inline keyboard with options like "10 min", "15 min", "30 min", "No timer") if they'd like to propose or agree to a timed session.
    *   **Agreement**: If one user proposes a time, the other user would be asked to agree. If both select the same duration, or one proposes and the other agrees, the timer starts. If there's no agreement (or one user selects "No timer"), the chat continues as an untimed session.
    *   **Notifications**: The bot would inform both users that the timed session has begun. It could also send reminders, for example, "5 minutes remaining" and "1 minute remaining."
*   **Functionality**:
    *   **Timer Expiry**: When the agreed-upon time expires, the bot sends a message to both users indicating the timed session has concluded.
    *   **Post-Timer Action**:
        *   The chat could be automatically disconnected (equivalent to both users issuing `/stop`).
        *   Alternatively, users could be given a brief option (e.g., via inline keyboard) to mutually agree to extend the timer (e.g., by another 5 or 10 minutes) or to continue untimed. If no agreement to extend is reached quickly, the chat would then disconnect.
*   **Anonymity Considerations**:
    *   This feature is primarily about session management and does not inherently pose significant anonymity risks if implemented carefully.
    *   User preferences for timed chats should ideally not be stored beyond the current session to avoid creating a persistent behavioral profile. The decision to use a timer should be made fresh for each chat session.
*   **Benefits**:
    *   **Structured Conversations**: Encourages more focused or goal-oriented conversations for users who prefer that.
    *   **Expectation Management**: Helps manage expectations regarding the potential duration of a chat.
    *   **Time Management**: Useful for users who have limited time or prefer shorter, time-boxed interactions.
    *   Can act as a natural end-point for conversations that might otherwise dwindle or end awkwardly.
*   **Complexity**:
    *   **Timer Management**: Requires robust server-side management of active timers for potentially many concurrent chat sessions. This would likely involve goroutines and tickers for each timed chat.
    *   **Agreement Logic**: Needs logic to handle the proposal and agreement flow between two users for setting or extending timers.
    *   **State Management**: The `User` or chat session object would need to store timer-related information (e.g., agreed duration, start time, timer active status).
    *   **Notifications**: Implementing the reminder and session-end notifications.

### 4. Optional 'Icebreaker' Prompts

*   **Concept**: After users are successfully connected for a 1-on-1 chat, the bot can offer to provide a random, neutral icebreaker question or conversation starter to both users simultaneously.
*   **User Experience**:
    *   **Optional Offer**: Once connected, the bot would privately send a message to each user, perhaps with an inline keyboard, asking something like: "You're connected! Want an icebreaker question to get things started? (Yes/No)".
    *   **Mutual Consent**: If *both* users select "Yes" (or a similar affirmative response) within a short timeframe, the bot then sends the *same* randomly selected icebreaker prompt to each of them.
    *   **Decline or No Response**: If either user selects "No", or if one user doesn't respond to the offer within a reasonable period (e.g., 15-30 seconds), no icebreaker is sent, and users are free to start the conversation on their own.
*   **Functionality**:
    *   **Predefined List**: Requires a curated, predefined list of appropriate, neutral, and engaging icebreaker questions/prompts (e.g., "If you could travel anywhere tomorrow, where would you go?", "What's a hobby you've always wanted to try?", "Coffee or tea?").
    *   **Consent Management**: Logic to manage the consent state from both users for receiving an icebreaker. This involves tracking responses to the offer.
    *   **Random Selection**: A mechanism to randomly select an icebreaker from the list if consent is given.
*   **Anonymity Considerations**:
    *   **Generic Prompts**: Icebreakers must be generic and open-ended, and must not solicit any personally identifiable information (PII) or overly private details.
    *   **Private Choice**: A user's choice to accept or decline an icebreaker prompt should be private and not revealed to their chat partner. Only the outcome (i.e., whether an icebreaker is sent to both or not) is implicitly known if an icebreaker appears.
*   **Benefits**:
    *   **Reduces Initial Awkwardness**: Helps users overcome the common hurdle of how to start a conversation with a complete stranger.
    *   **Provides Shared Starting Point**: Gives both users a clear, neutral topic to begin their interaction, making the initial engagement smoother.
    *   **Encourages Interaction**: Can prompt users who might otherwise be hesitant to send the first message.
*   **Complexity**:
    *   **Consent Flow Management**: Implementing the logic to offer the icebreaker, track responses from both users, and act accordingly (send prompt or do nothing).
    *   **Content Curation**: Developing and maintaining a good, diverse list of suitable icebreaker questions that are engaging but also safe and neutral.
    *   **Minor Post-Connection Logic**: Adds a step to the bot's actions immediately after a successful connection.

### 5. Karma/Trust System (Anonymous & Simplified) - **HIGHLY EXPERIMENTAL & SENSITIVE**

**VERY STRONG CAVEATS ON ANONYMITY AND COMPLEXITY:**

*   **EXTREME DIFFICULTY AND RISK**: This feature suggestion is by far the **MOST COMPLEX, SENSITIVE, AND POTENTIALLY RISKY** of all ideas presented. It should be approached with extreme caution, if at all. The potential for unintended consequences or failure to maintain perfect anonymity is very high.
*   **ANONYMITY RISKS ARE PARAMOUNT**: Implementing *any* form of reputation or karma system without compromising user anonymity, even indirectly, is **EXCEPTIONALLY CHALLENGING**. Even aggregated, non-identifiable data could potentially be misused, misinterpreted, or lead to emergent behaviors that undermine the anonymous nature of the bot. **FAILURE TO MAINTAIN ANONYMITY WOULD BE A CRITICAL FAILURE OF THE BOT'S CORE PURPOSE.**
*   **NO DIRECT USER-TO-USER RATING**: It is crucial that users **DO NOT** rate each other directly in a way that links one user's specific rating to another specific user's identity or Telegram ID. Any feedback mechanism must be about the *general chat experience* and submitted *after* the chat has ended and both users have disconnected from each other.

*   **Concept**: A highly simplified, anonymous system allowing users to provide minimal feedback (e.g., "good chat experience" / "bad chat experience") *after* a chat has ended and they are fully disconnected from their partner. The primary goal is to very subtly discourage consistently problematic behavior, not to build visible reputations.
*   **Potential Mechanics (Highly Abstract & Simplified)**:
    *   After a user issues `/stop` and the chat is confirmed as ended, the bot could privately ask the user: "Did you have a positive chat experience this time? (Yes/No)".
    *   This feedback is recorded anonymously by the system. **It is NEVER shown to the other user involved in the chat.**
    *   **VERY CAUTIOUSLY**, the system *might* use heavily aggregated, anonymous data from *many different, unrelated* chat sessions over extended periods.
        *   If a user consistently receives a very high ratio of "negative experience" feedback from numerous distinct partners, this *might* (when combined with other indicators like verified spam reports, if such a system existed) slightly deprioritize them in the matching queue for a limited time. This is about identifying *patterns* of consistently problematic behavior, not penalizing based on isolated incidents or individual user ratings.
        *   Conversely, consistently high positive feedback from many distinct partners *might* subtly, and very slightly, increase a user's chance of being matched, but **NEVER** in a way that creates "elite" user tiers or significantly alters the random nature of matching.
*   **Purpose**: Primarily to discourage widespread, repeated abuse, spam, or consistently negative behavior that harms the platform's health, rather than to build any form of detailed or visible user reputation.
*   **Anonymity Safeguards (Essential & Non-Negotiable)**:
    *   User-provided feedback is **NEVER** visible to other users or the partner from the chat.
    *   There must be **NO** leaderboards, visible karma scores, or any user-facing indication of their own or others' aggregated feedback.
    *   Any internal metrics based on this feedback must be heavily aggregated, anonymized (e.g., using statistical noise or differential privacy techniques if feasible), and time-decayed (older feedback becomes less relevant).
    *   The system's primary focus must be on identifying broad *patterns* of abuse, not on "rating" individual users for general matching purposes.
    *   **Transparency with users about the *existence* of such an internal, anonymized system (if implemented) would be important, while keeping its exact mechanics opaque to prevent gaming.**
*   **Benefits (Tentative & Uncertain)**:
    *   Could *potentially* and *subtly* improve the overall quality of chat experiences by discouraging users who consistently engage in disruptive, offensive, or spammy behavior.
    *   Might provide a very basic, anonymized signal for identifying users who frequently violate community standards (if such standards were clearly defined).
*   **HIGH COMPLEXITY & RISK REITERATED**:
    *   This feature is **EXTREMELY HIGH-RISK AND HIGH-COMPLEXITY**.
    *   The technical challenges in ensuring true anonymity while still deriving a useful signal are immense.
    *   The potential for misinterpretation of data, creating unfair biases, or users attempting to "game" the system is significant.
    *   **It may very well NOT be suitable or worth the risks and development effort for an anonymous chat bot.** Thorough ethical review and technical scrutiny would be paramount before even considering a prototype.

## Feature Details

This section provides more in-depth explanations of how specific complex features, particularly those involving new user commands or interactions, are envisioned to work.

### Interest-Based Matching

Interest-based matching is a proposed feature (see "Future Enhancements") to allow users to optionally select a few predefined interests to help the bot pair them with others who share common ground. This feature is designed with anonymity as a top priority.

**Command: `/myinterests`**

*   **Purpose**: The `/myinterests` command allows users to manage their anonymous interest selections. These selections are used by the bot to *attempt* to find more compatible chat partners. Importantly, a user's selected interests are **never directly revealed** to their chat partner. The goal is to make connections feel more natural if common ground exists, not to explicitly state shared interests.

*   **Predefined Interests**: Users can choose from a predefined list of 8 broad interest categories. This list is fixed to help maintain anonymity and ensure a sufficiently large pool for each interest:
    1.  `Movies & TV`
    2.  `Music`
    3.  `Gaming`
    4.  `Books & Writing`
    5.  `Sports & Fitness`
    6.  `Tech & Science`
    7.  `Travel & Outdoors`
    8.  `Food & Cooking`

*   **User Interface Flow for `/myinterests`**:

    When a user sends the `/myinterests` command, the bot responds with:

    1.  **Initial Message & Main Menu**:
        *   Text: "Manage your anonymous interests. Selecting interests can help us find more compatible chat partners for you. You can select up to 3 interests. Your interests are not directly shared with your chat partners."
        *   Inline Keyboard Buttons:
            *   `[ View/Edit My Interests ]`
            *   `[ Clear My Interests ]`
            *   `[ Cancel ]`

    2.  **If "View/Edit My Interests" is selected**:
        *   The bot displays a message showing the user's currently selected interests (if any). For example:
            *   "Your current interests: Music, Gaming. You can select up to 3."
            *   Or, if none: "You currently have no interests selected. You can select up to 3."
        *   Below this message, an inline keyboard is shown with all 8 predefined interests.
            *   Each interest button will indicate its current selection status. For example, a selected interest might have a "✅" prepended or appended (e.g., `[ ✅ Music ]`), while an unselected one does not (e.g., `[ Gaming ]`).
            *   A "Done Editing" button is also displayed.
        *   **Interaction Logic**:
            *   Tapping an unselected interest button selects it (and updates the ✅).
            *   Tapping a selected interest button deselects it.
            *   If the user tries to select a 4th interest, the bot sends a brief message like: "You can select a maximum of 3 interests. Please deselect one if you wish to choose another." The selection is not made.
            *   The inline keyboard is updated after each tap to reflect the current selections.
        *   When the user taps **`[ Done Editing ]`**:
            *   The bot saves the current selections (anonymously, associated with their user ID for the current session/until cleared).
            *   It sends a confirmation message: "Your interests have been updated to: [List of selected interests]." or "Your interests have been saved. You currently have [X] interests selected." If no interests are selected, it might say: "Your interests have been saved. You currently have no interests selected."
            *   The `/myinterests` interaction ends.

    3.  **If "Clear My Interests" is selected**:
        *   The bot clears any previously selected interests for the user.
        *   It sends a confirmation message: "Your anonymous interests have been cleared."
        *   The `/myinterests` interaction ends.

    4.  **If "Cancel" is selected**:
        *   The bot sends a message: "No changes made to your interests."
        *   The `/myinterests` interaction ends.

*   **Anonymity Reminder**: The system is designed such that these selected interests are used as a preference for matching but are not explicitly disclosed to chat partners. The aim is for conversations to potentially feel more natural due to shared (but unstated) common ground.
