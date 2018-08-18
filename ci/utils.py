from os import system

def execute_commands(cmds, fail=True, print_cmd=True):
    for cmd in cmds:
        if print_cmd:
            print(f'>>> {cmd}')
        else:
            print(f'>>> ***')
        code = system(cmd)
        if code < 0 and fail:
            exit(code)
