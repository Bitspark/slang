from os import system

def execute_commands(cmds, fail=True):
    for cmd in cmds:
        print(f'>>> {cmd}')
        code = system(cmd)
        if code < 0 and fail:
            exit(code)
