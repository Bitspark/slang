"""Run before the public router; only the privileged agent may inspect Docker."""
from runtime_agent import reconcile_all

if __name__ == '__main__':
    reconcile_all()
