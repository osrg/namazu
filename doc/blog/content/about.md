+++
categories = ["general"]
date = "2016-04-27"
tags = ["document"]
title = "About Namazu"

+++

In short, the goal of Namazu project is providing a foundation of debugger for distributed systems.

Developing and maintaining distributed systems is difficult. 
The difficulty comes from many factors, 
but we believe that one of the most important reasons is lacking of a good debugger for distributed systems specific bugs.

What are the distributed systems specific bugs?
The bugs caused by hardware faults, non determinism of message ordering, and mix of them
(non distributed systems does not need to handle them).
Some researchers point out that real world systems (e.g. Hadoop) have such bugs, 
which can cause terrible failures like data loss [\[1\]][1] [\[3\]][3]. 
They showed the existence of the bugs by building implementation level distributed system model checkers (DMCK).
The DMCKs have a capability of searching complex state space of protocols and injeting faults at critical timings.
In addition, they can work with actual implementation (not formal model [\[2\]][2]) directly.

Namazu is a programmable fuzz testing framework inspired by such DMCKs.
Its design does not depend on programming languages and opearting systems.
You can write your own state space search policy for your system.
We hope it will make your life a little bit easier.

![Overview](/namazu/images/namazu.png)

[1]: https://www.usenix.org/conference/osdi14/technical-sessions/presentation/leesatapornwongsa "Tanakorn Leesatapornwongsa, et al. SAMC: Semantic-Aware Model Checking for Fast Discovery of Deep Bugs in Cloud Systems. In Proc. of OSDI '14."
[2]: http://research.microsoft.com/en-us/um/people/lamport/tla/formal-methods-amazon.pdf "Chris Newcombe, et al. Use of Formal Methods at Amazon Web Services. Amazon.com Technical Report, 2014."
[3]: https://www.usenix.org/legacy/event/nsdi09/tech/full_papers/yang/yang.pdf "Junfeng Yang, et al. MODIST: Transparent Model Checking of Unmodified Distributed Systems. In Proc. of NSDI '09."

## References
1. [Tanakorn Leesatapornwongsa, et al. SAMC: Semantic-Aware Model Checking for Fast Discovery of Deep Bugs in Cloud Systems. In Proc. of OSDI '14.][1]
2. [Chris Newcombe, et al. Use of Formal Methods at Amazon Web Services. Amazon.com Technical Report, 2014.][2]
3. [Junfeng Yang, et al. MODIST: Transparent Model Checking of Unmodified Distributed Systems. In Proc. of NSDI '09.][3]
