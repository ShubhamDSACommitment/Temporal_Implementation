When you run go run cmd/starter/main.go, here's what happens:

1. Workflow starts (2 actions)

Starter calls ExecuteWorkflow()                                                                                                                                                                                  
→ Action: WorkflowStarted                               
→ Action: WorkflowTaskStarted (Temporal schedules the workflow code to run)

2. ValidateOrder activity (4 actions)

// order_workflow.go:50                                                                                                                                                                                          
err := workflow.ExecuteActivity(orderCtx, activities.ValidateOrder, order).Get(ctx, &validationStatus)                                                                                                           
→ Action: ActivityTaskScheduled     (Temporal puts task on order-task-queue)                                                                                                                                     
→ Action: ActivityTaskStarted       (order-service worker picks it up)                                                                                                                                           
→ Action: ActivityTaskCompleted     (worker returns "VALIDATED")                                                                                                                                                 
→ Action: WorkflowTaskStarted      (workflow code resumes to execute next line)

3. ChargePayment — the cross-service call (4 actions)

// order_workflow.go:65                                                                                                                                                                                          
err = workflow.ExecuteActivity(paymentCtx, "ChargePayment", paymentReq).Get(ctx, &payment)                                                                                                                       
→ Action: ActivityTaskScheduled     (Temporal puts task on payment-task-queue)                                                                                                                                   
→ Action: ActivityTaskStarted       (payment-service worker picks it up)      
→ Action: ActivityTaskCompleted     (worker returns PaymentResult)                                                                                                                                               
→ Action: WorkflowTaskStarted      (workflow code resumes)

Same number of actions whether it's local or cross-service. Temporal doesn't care.

4. ShipOrder (4 actions)

// order_workflow.go:72                                   
err = workflow.ExecuteActivity(orderCtx, activities.ShipOrder, order).Get(ctx, &shipment)                                                                                                                        
→ Action: ActivityTaskScheduled                                                                                                                                                                                  
→ Action: ActivityTaskStarted                                                                                                                                                                                    
→ Action: ActivityTaskCompleted                                                                                                                                                                                  
→ Action: WorkflowTaskStarted

5. SendNotification (4 actions)

// order_workflow.go:80                                   
err = workflow.ExecuteActivity(orderCtx, activities.SendNotification, order.OrderID, message).Get(ctx, nil)                                                                                                      
→ Action: ActivityTaskScheduled                                                                                                                                                                                  
→ Action: ActivityTaskStarted  
→ Action: ActivityTaskCompleted                                                                                                                                                                                  
→ Action: WorkflowTaskStarted

6. Workflow completes (1 action)

→ Action: WorkflowCompleted

Total Count

┌────────────────────────┬─────────┐                      
│          Step          │ Actions │
├────────────────────────┼─────────┤                                                                                                                                                                             
│ Workflow start         │ 2       │
├────────────────────────┼─────────┤                                                                                                                                                                             
│ ValidateOrder          │ 4       │                      
├────────────────────────┼─────────┤                                                                                                                                                                             
│ ChargePayment          │ 4       │
├────────────────────────┼─────────┤                                                                                                                                                                             
│ ShipOrder              │ 4       │                      
├────────────────────────┼─────────┤
│ SendNotification       │ 4       │
├────────────────────────┼─────────┤                                                                                                                                                                             
│ Workflow complete      │ 1       │
├────────────────────────┼─────────┤                                                                                                                                                                             
│ Total per workflow run │ ~19     │                      
└────────────────────────┴─────────┘

What About Retries?

If ChargePayment fails and Temporal retries it (you set MaximumAttempts: 3):

First attempt:   ActivityTaskScheduled + ActivityTaskStarted + ActivityTaskFailed  = 3 actions                                                                                                                   
Second attempt:  ActivityTaskScheduled + ActivityTaskStarted + ActivityTaskCompleted = 3 actions
+ WorkflowTaskStarted = 1 action

Each retry adds ~3 extra actions.

Scaling This to Your Org

19 actions × 1,000 orders/day   = 19,000 actions/day                                                                                                                                                             
19 actions × 100,000 orders/day = 1.9M actions/day → ~57M/month

At $25 per 1M actions:                                                                                                                                                                                           
1,000 orders/day  → ~$15/month                                                                                                                                                                                 
100,000 orders/day → ~$1,425/month

That's just the order workflow. Multiply by however many workflows your other products would run.

Simple Rule of Thumb

Each activity = ~4 actions. So count your activities, multiply by 4, add 3 for workflow overhead. That's your per-execution cost. 