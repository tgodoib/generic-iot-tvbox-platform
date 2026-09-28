Import("env")

with open("VERSION") as f:
    version = f.read().strip()

print(f"[pre_build] SSAIOT_FW_VERSION = {version}")
env.Append(CPPDEFINES=[("SSAIOT_FW_VERSION", f'\\"{version}\\"')])