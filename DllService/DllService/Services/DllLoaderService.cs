using System;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Collections.Generic;

public class DllLoaderService
{
    private readonly Dictionary<string, Assembly> _loadedAssemblies = new Dictionary<string, Assembly>();

    public void LoadDllsFromDirectories(IEnumerable<string> directories)
    {
        foreach (var dir in directories)
        {
            var assembly = Assembly.LoadFrom(dir);
            _loadedAssemblies[dir] = assembly;
        }
    }

    /// <summary>
    /// Get the method from DLL
    /// </summary>
    /// <param name="dllPath"></param>
    /// <param name="typeName"></param>
    /// <param name="methodName"></param>
    /// <returns></returns>
    public MethodInfo? GetMethod(string dllPath, string typeName, string methodName)
    {
        if (_loadedAssemblies.TryGetValue(dllPath, out var assembly))
        {
            var type = assembly.GetType(typeName);
            return type?.GetMethod(methodName);
        }
        return null;
    }

    /// <summary>
    /// Get the parameters of the method from DLL
    /// </summary>
    /// <param name="method"></param>
    public object[] GetParameters(MethodInfo method, bool initializeProperties = true)
    {
        // ���omethod���Ѽ�Info
        var parametersInfo = method.GetParameters();
        if (parametersInfo.Length == 0)
        {
            throw new ArgumentException("�L�k���omethod���Ѽ�Info");
        }

        // �ŧi�Ѽư}�C�A�Ѽƥi��O�h��
        var parameters = new object[parametersInfo.Length];

        for (int i = 0; i < parametersInfo.Length; i++)
        {
            // �إ߰Ѽ��ݩ�
            var paramType = parametersInfo[i].ParameterType;

            // �إ߰Ѽƹ��
            var paramInstance = Activator.CreateInstance(paramType);

            if (paramInstance == null)
            {
                if (paramType == typeof(string))
                {
                    paramInstance = string.Empty; // �q�{�Ŧr��
                }
                else if (paramType.IsInterface)
                {
                    throw new NotSupportedException($"Cannot initialize interface type: {paramType}");
                }
                else
                {
                    throw new ArgumentException($"Cannot initialize parameter of type {paramType}");
                }
            }

            if (initializeProperties)
            {
                // ��l���ݩʡA�����]�w��
                foreach (var property in paramType.GetProperties())
                {
                    if (property.CanWrite) // �T�O�ݩʬO�i�g��
                    {
                        if (property.PropertyType.IsValueType)
                        {
                            // ���������]�m�q�{�� (0�Bfalse ��)
                            property.SetValue(paramInstance, Activator.CreateInstance(property.PropertyType));
                        }
                        else if (property.PropertyType.IsClass)
                        {
                            // ���ޥ������]�m�� null �ΪŹ�ҡ]�p�G�����i��l�ơ^
                            property.SetValue(paramInstance, null);
                        }
                    }
                }
            }
            parameters[i] = paramInstance;
        }

        return parameters;
    }

    /// <summary>
    /// Invoke the method from DLL
    /// </summary>
    /// <param name="dllPath"></param>
    /// <param name="typeName"></param>
    /// <param name="methodName"></param>
    /// <param name="parameters"></param>
    /// <returns></returns>
    /// <exception cref="ArgumentException"></exception>
    public object? InvokeMethod(string dllPath, string typeName, string methodName, params object[] parameters)
    {
        var method = GetMethod(dllPath, typeName, methodName);
        if (method != null)
        {
            var parametersInfo = method.GetParameters();
            if (parametersInfo.Length != parameters.Length)
            {
                throw new ArgumentException("The number of parameters does not match the method signature.");
            }

            for (int i = 0; i < parametersInfo.Length; i++)
            {
                if (parameters[i] != null && !parametersInfo[i].ParameterType.IsAssignableFrom(parameters[i].GetType()))
                {
                    throw new ArgumentException($"Parameter at index {i} is not compatible with the expected type {parametersInfo[i].ParameterType}.");
                }
            }

            var instance = method.IsStatic ? null : Activator.CreateInstance(method.DeclaringType!);
            return method.Invoke(instance, parameters);
        }
        return null;
    }



}