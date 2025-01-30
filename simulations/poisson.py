import csp
from csp import ts
from datetime import timedelta
import numpy as np

@csp.node(memoize=False)
def poisson_counter(rate: float) -> ts[int]:
    with csp.alarms():
        event = csp