import pdb
from .. import LOG as _LOG

LOG = _LOG.getChild('util.util')


class Breakpoint(object):
    DISABLE_BP = False

    @classmethod
    def bp(cls):
        if cls.DISABLE_BP:
            return
        LOG.debug('Dropping to PDB.. (type "c" to continue)')
        pdb.set_trace()
