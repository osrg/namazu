+++
categories = ["blog"]
date = "2015-07-24"
tags = ["document"]
title = "Hello, world!"

+++


Hello, world! In this blog, we'd like to share our ideas and
experiences related to testing and debugging distributed systems.

Testing and debugging software is difficult. Especially, testing and
debugging distributed systems is known to be *very* difficult. Why so
difficult? We believe the difficulty comes from lacking good tools for
the distributed systems specific bugs.

Testing, debugging, and verification techniques for
removing bugs in software have a long history. Even programming is difficult task
since the beginning of its history, significant effort of researchers and
engineers is succeeding at establishing techniques for fighting
against major important bugs. For example, modern programming
languages tend to have their own GC mechanisms and they are very
effective for removing memory leak bugs. In addition, tools like
valgrind can help programmers to detect these bugs even software is
written in C or C++. Many other techniques were also established for
other types of bugs, so developing software seems to be becoming
easier than ancient days.

However, times are changing. In these days, it is clear that
importance of a new category of software, distributed systems, is
rising. Although the idea of distributed systems is very old, its
modern implementations, e.g. Apache Hadoop, are introducing
significant benefit to today's world. Distributed systems are
essentially different from non-distributed systems. They combine
multiple computers for highly availabile, highly durabile, and
scalable performant systems. As a result, they can enable new sort of
services called cloud computing and big data. Therefore everyone loves
these systems. 

Sadly, [recent studies](http://ucare.cs.uchicago.edu/pdf/socc14-cbs.pdf) shows that bugs in
the distributed systems are hard to detect and tend to introduce critical
failures e.g. permanent data loss. Of course distributed systems share
many types of bugs (e.g. memory leak, race conditions) with
non-distributed systems, so existing debugging
techniques are also effective for removing these bugs in the
systems. However, distributed systems have their own types of bugs and
the critical failures tend to be introduced by such bugs.

What are the distributed systems specific bugs? Studies of this area
is in very early stage so there's no mature categorization, but I can list
some classes which can be seen in many systems:

 * distributed race conditions caused by interleaving of messages over network
 * incorrect handling of hardware e.g. disk failure, node failure, network partition
 * performance degrading, especially a case of losing scalability

Of course all of the above three classes are critical, but first and
second ones are especially emergent because they are related to
correctness of systems. Incorrect systems are not valuable even
their performance is good.

Though programming methodologies are evolving, these bugs remain hard
to be debugged because of lacking good tools. In succeeding posts, we
will describe the difficulties of the debugging, possible candidates
of solutions, and a tool we are working on.
