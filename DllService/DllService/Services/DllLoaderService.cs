using System;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Collections.Generic;

public class DllLoaderService
{
    private readonly Dictionary<string,Assembly> _loadedAssemblies = new Dictionary<string, Assembly>();

    public void LoadDllsFromDirectories(IEnumerable<string> directories)
    {
        foreach (var dir in directories)
        {
            if (Directory.Exists(dir))
            {
                var dllFiles = Directory.GetFiles(dir,"*.dll");
                foreach (var file in dllFiles)
                {
                    var assembly = Assembly.LoadFrom(file);
                    _loadedAssemblies[file] = assembly;
                }
            }
        }
    }

    public MethodInfo? GetMethod(string dllPath,string typeName,string methodName)
    {
        if (_loadedAssemblies.TryGetValue(dllPath, out var assembly))
        {
            var type = assembly.GetType(typeName);
            return type?.GetMethod(methodName);
        }
        return null;
    }

    public object? InvokeMethod(string dllPath, string typeName,string methodName, object[] parameters)
    {
        var method = GetMethod(dllPath,typeName,methodName);
        if (method != null)
        {
            var instance = Activator.CreateInstance(method.DeclaringType!);
            return method.Invoke(instance, parameters);
        }
        return null;
    }
}