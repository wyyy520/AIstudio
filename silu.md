AIStudio 系统设计说明书（System Design Document）
第一章 项目背景与总体架构设计
1.1 项目背景
随着大语言模型（LLM）、AI Agent 以及生成式 AI 的快速发展，越来越多的开发工作开始依赖 AI 辅助完成。当前市场上已经出现了 Cursor、Claude Code、GitHub Copilot 等智能开发工具，它们能够根据自然语言生成代码，提高开发效率。然而，这类工具本质上仍然属于"代码生成工具"，开发者依旧需要理解工程结构、维护项目架构以及处理大量工程配置，对于人工智能工程、MATLAB 仿真、STM32 嵌入式开发、ANSYS 仿真等专业工程领域来说，开发门槛依然较高。
另一方面，专业软件之间缺乏统一的开发模式。Python 负责 AI 训练，MATLAB 负责算法仿真，STM32CubeMX 用于嵌入式工程生成，ANSYS 用于有限元分析，各个平台之间相互独立，开发者需要频繁切换软件，并手动完成数据转换、工程创建以及环境配置，整个开发流程复杂且重复劳动较多。
AIStudio 的目标并不是继续优化代码生成，而是重新思考整个工程开发流程，希望以"工作流"代替"代码"，以"工程生成"代替"工程搭建"，最终形成一套统一的工程开发方式，使不同领域的软件能够使用同一套工作流进行描述，并自动生成对应平台的真实工程。
1.2 项目定位
AIStudio 是一款面向 AI 工程与专业工程开发的可视化低代码开发平台。
它不是聊天机器人。
它不是代码补全工具。
它不是传统意义上的 AI IDE。
它也不是某一个 AI 模型训练平台。
AIStudio 的定位是：
以 Workflow 描述工程，以 Compiler 编译工程，以 Template 生成工程，以 Runtime 运行工程，以 AI 增强工程。
用户在平台中无需直接编写大量代码，而是通过拖拽节点、连接工作流、配置节点参数完成整个工程设计，系统自动完成工程生成、环境调用以及运行调试。
平台支持人工智能训练、MATLAB 仿真、STM32 嵌入式开发、ANSYS 仿真分析等多个专业领域，并采用统一的数据结构进行描述，实现跨平台、跨领域、可扩展的工程开发。
1.3 项目目标
AIStudio 希望解决以下几个核心问题：
第一，降低 AI 工程及专业工程开发门槛，使开发者更多关注业务逻辑，而不是大量重复的工程配置。
第二，建立统一的工作流开发模式，不同软件均采用 Workflow 描述工程，而不是各自维护独立配置方式。
第三，建立标准工程生成体系，由系统自动生成 Python、MATLAB、STM32、ANSYS 等真实工程，而非生成零散代码片段。
第四，实现 AI 与工程开发解耦，大语言模型仅作为辅助能力，不参与核心工程生成流程，保证系统离线可运行、可维护。
第五，构建开放式插件生态，使第三方开发者能够通过插件扩展新的节点、模板、Generator、Skill 以及专业软件支持。
最终，希望 AIStudio 能够成为一个统一的工程开发平台，而不仅仅是一个 AI 工具。
1.4 核心设计理念
AIStudio 的整个系统围绕五个核心理念进行设计。
第一，Workflow First。
开发者开发的对象不再是代码，而是工作流。
整个工程均由工作流描述。
所有 Generator 均根据工作流生成工程。
工作流永远是真实数据来源。
第二，JSON IR（Intermediate Representation）。
Workflow 不直接生成代码，而是统一转换为标准 Workflow JSON。
Workflow JSON 不仅仅是配置文件，更是整个 AIStudio 的工程中间表示（IR）。
所有 Compiler、Generator、Runtime 均围绕 Workflow JSON 工作。
第三，Compiler Driven。
AIStudio 引入 Compiler 作为整个系统核心。
Compiler 负责解析 Workflow、校验工作流、分析节点依赖、生成 Execution Plan，并调用不同 Generator。
Compiler 不负责代码生成。
Compiler 负责工程编译。
第四，Template Driven。
平台不直接拼接代码。
所有工程均基于标准模板生成。
不同领域拥有不同工程模板。
例如：
Python 使用 Python Template。
MATLAB 使用 MATLAB Template。
STM32 使用 CubeMX Template。
ANSYS 使用 Journal Template。
Compiler 仅负责将 Workflow 参数填充至模板，最终生成真实工程。
第五，AI Optional。
AI 并不是平台核心。
AIStudio 即使关闭所有 LLM，也能够完成：
工作流编辑。
工程保存。
工程生成。
工程运行。
AI 仅承担：
工作流规划。
日志分析。
参数解释。
错误诊断。
优化建议。
等辅助能力。
1.5 总体架构
AIStudio 采用模块化分层架构。
整个系统由以下几个核心模块组成：
Project Manager 负责项目生命周期管理，包括创建项目、打开项目、保存项目以及项目目录维护。
Workflow System 负责可视化节点编辑、节点连接、属性配置以及 Workflow 数据维护。
Workflow Store 负责维护整个工作流数据，并实时同步 Workflow JSON。
Compiler 负责解析 Workflow JSON，完成工作流合法性校验、拓扑排序、依赖分析，并生成 Execution Plan。
Template Engine 负责管理所有工程模板，提供模板复制、变量替换以及模板扩展能力。
Generator 根据 Execution Plan 调用对应模板，生成 Python、MATLAB、STM32、ANSYS 等真实工程。
Runtime 负责启动本地开发环境，运行生成后的工程，管理运行状态以及终端输出。
Log Center 统一管理所有运行日志，包括标准输出、错误输出、环境信息以及运行历史。
Diagnose Center 负责日志解析、错误分析、环境检查以及自动修复建议。
Skill Center 提供 AI 能力，包括 Workflow Planner、Explain、Diagnose、Optimize 等 AI 服务。
Plugin Center 提供插件生态，支持节点插件、Generator 插件、模板插件、Skill 插件以及 Provider 插件。
整个系统各模块之间相互独立，仅通过统一数据接口进行通信，保证平台具备良好的可维护性和扩展能力。
1.6 系统数据流
AIStudio 整个数据流采用统一工作流驱动方式。
开发者首先创建项目。
随后在 Workflow Editor 中拖拽节点。
配置节点参数。
连接节点端口。
Workflow Store 实时更新。
系统自动同步生成 workflow.json。
Compiler 读取 workflow.json。
解析节点。
分析连接关系。
生成 Execution Plan。
Generator 根据 Execution Plan 调用对应 Template。
Template Engine 完成变量替换。
生成真实工程。
Runtime 自动调用本地开发环境运行工程。
运行日志统一进入 Log Center。
Diagnose Center 对日志进行分析。
若开启 AI，则 Skill Center 调用大语言模型进行错误解释、参数优化以及工作流建议。
整个过程形成完整闭环。
1.7 系统最终目标
AIStudio 最终希望建立一种全新的工程开发模式。
开发者不再围绕代码开发。
而是围绕工程工作流开发。
代码只是工程生成后的产物。
Workflow 才是真正的工程描述。
整个系统最终形成如下开发流程：
业务需求 → Workflow → Workflow JSON → Compiler → Execution Plan → Template → Generator → Real Project → Runtime → Log Center → Diagnose → AI(Optional)
这一流程将作为 AIStudio 后续所有模块设计、插件开发以及功能扩展的统一标准，也是整个项目长期演进过程中始终遵循的核心架构思想。





第二章 Workflow 系统设计

2.1 Workflow 设计背景

Workflow（工作流）是 AIStudio 的核心，也是整个系统唯一的数据来源（Single Source of Truth）。

传统 AI 开发流程通常需要开发者手动创建 Python 工程、配置数据集、编写训练代码、维护工程目录，并在多个软件之间不断切换。而 AIStudio 希望将整个开发过程抽象为一张可视化工作流，将原本复杂的软件工程流程转换为节点（Node）与连线（Edge）的组合，使开发者能够像搭建流程图一样完成整个工程设计。

因此，在 AIStudio 中，用户真正编辑的对象不是 Python 代码，也不是 MATLAB 脚本，而是 Workflow。任何代码、任何工程、任何配置均由 Workflow 自动生成。

2.2 Workflow 在系统中的定位

Workflow 是整个 AIStudio 的最上层描述。

整个系统所有模块均围绕 Workflow 工作。

任何模块不得直接读取前端状态。

任何模块不得直接操作页面组件。

整个系统的数据流必须遵循：

Workflow Editor │ ▼ Workflow Store │ ▼ workflow.json │ ▼ Compiler │ ▼ Execution Plan │ ▼ Generator │ ▼ Runtime 

因此，Workflow 不仅负责界面显示，更承担整个工程描述的职责。

2.3 Workflow 编辑器

Workflow Editor 是用户唯一的工程编辑入口。

用户可以通过拖拽方式创建整个工程，而无需直接编写代码。

Workflow Editor 至少需要支持以下功能：

创建节点。

删除节点。

复制节点。

粘贴节点。

移动节点。

缩放画布。

框选节点。

多选节点。

撤销。

重做。

自动排列。

节点搜索。

节点分类。

节点收藏。

快捷键操作。

小地图（MiniMap）。

网格吸附。

自动对齐。

所有操作均实时同步至 Workflow Store。

2.4 Node（节点）设计

节点是 Workflow 中最小的业务单元。

每一个节点代表一种工程能力，而不是一段代码。

例如：

Dataset 节点代表数据集。

YOLO 节点代表目标检测训练。

LSTM 节点代表时序预测。

MATLAB 节点代表 MATLAB 仿真。

STM32 节点代表嵌入式工程。

ANSYS 节点代表有限元分析。

Python Script 节点代表 Python 自定义脚本。

Node 的职责是描述能力，而不是执行能力。

真正执行由 Generator 完成。

2.5 Node 数据结构

每一个 Node 必须拥有统一的数据结构。

包括：

唯一 ID。

节点类型。

节点名称。

节点位置。

节点尺寸。

输入端口。

输出端口。

参数配置。

节点状态。

启用状态。

创建时间。

更新时间。

扩展字段。

Plugin 信息。

Domain 信息。

Node 必须支持未来插件动态扩展。

任何插件均可向 Node 增加新的属性，而无需修改 Workflow Schema。

2.6 Edge（连线）设计

Edge 用于描述节点之间的数据流与依赖关系。

Edge 并不是一条普通连线，而是 Workflow 中真正决定执行顺序的重要数据。

每一条 Edge 至少应包含：

唯一 ID。

起始节点。

目标节点。

起始端口。

目标端口。

连接类型。

连接标签。

条件表达式。

未来扩展字段。

Compiler 将根据 Edge 自动完成拓扑排序，并生成最终执行计划。

2.7 Property Panel（属性面板）

Workflow 中每一个节点均拥有独立属性面板。

用户点击节点后，右侧自动显示对应节点参数。

例如：

YOLO 节点显示：

模型名称。

Epoch。

Batch Size。

Image Size。

Workers。

Device。

Optimizer。

Learning Rate。

数据集路径。

输出目录。

所有参数修改后必须实时同步至 Workflow Store。

禁止使用"保存参数"按钮。

整个系统采用实时数据绑定。

2.8 Workflow Store

Workflow Store 是整个 Workflow 的唯一状态管理中心。

Workflow Editor 仅负责展示。

任何数据修改均通过 Workflow Store 完成。

Workflow Store 负责：

节点新增。

节点删除。

节点移动。

节点参数修改。

节点启用。

节点禁用。

节点复制。

节点粘贴。

连线新增。

连线删除。

Viewport 更新。

自动保存。

撤销。

重做。

所有页面均不得维护自己的 Workflow 状态。

2.9 workflow.json

Workflow Store 的所有数据最终实时同步至 workflow.json。

workflow.json 是整个系统唯一事实源。

任何 Generator。

任何 Runtime。

任何 Compiler。

任何 Skill。

任何 Plugin。

均只能读取 workflow.json。

不得读取 Vue Store。

不得读取组件状态。

workflow.json 不仅保存节点，还保存：

项目信息。

工作流版本。

插件信息。

Domain 信息。

节点信息。

连线信息。

画布信息。

变量信息。

元数据。

未来所有工程均基于 workflow.json 自动生成。

2.10 Viewport（画布）

Viewport 用于保存整个画布状态。

包括：

当前缩放比例。

画布偏移量。

当前中心点。

当前选择节点。

MiniMap 状态。

网格状态。

吸附状态。

重新打开项目后，应恢复与关闭前完全一致的画布状态。

2.11 Project Manager

Project Manager 负责整个 Workflow 生命周期。

包括：

创建项目。

打开项目。

保存项目。

另存项目。

关闭项目。

最近项目。

项目历史。

项目目录。

Project Manager 必须直接操作真实文件夹，而不是虚拟目录。

项目目录建议如下：

Project │ ├ workflow.json ├ project.json ├ templates/ ├ generated/ ├ datasets/ ├ outputs/ ├ logs/ ├ cache/ └ plugins/ 

2.12 Workflow 校验

Workflow 在保存之前必须完成合法性检查。

包括：

是否存在孤立节点。

是否存在循环依赖。

节点参数是否合法。

节点是否缺少输入。

输出端口是否重复连接。

节点版本是否兼容。

插件是否存在。

Template 是否存在。

Generator 是否存在。

所有错误均应在 Workflow 阶段提示，而不是等 Generator 报错。

2.13 Workflow 与 Compiler 的关系

Workflow 不负责执行。

Workflow 不负责生成代码。

Workflow 不负责运行工程。

Workflow 的唯一职责是描述工程。

Compiler 是 Workflow 的唯一消费者。

Workflow 输出：

workflow.json。

Compiler 输入：

workflow.json。

Workflow 不允许直接调用 Generator。

Workflow 不允许直接调用 Runtime。

整个系统必须严格按照：

Workflow → Compiler → Generator

执行。

2.14 Workflow 与 AI 的关系

AI 并不是 Workflow 的必要组成部分。

Workflow 可以完全脱离 AI 独立运行。

AI 在 Workflow 中仅提供辅助能力，例如：

根据自然语言自动生成工作流。

根据需求推荐节点。

自动完成节点连接。

参数推荐。

工作流优化建议。

节点解释。

关闭 AI 后：

用户依然能够：

创建 Workflow。

编辑 Workflow。

保存 Workflow。

生成工程。

因此，Workflow 必须与 AI 完全解耦。

2.15 Workflow 的最终目标

Workflow 是 AIStudio 最重要的基础设施，也是整个平台的核心。

未来，无论支持 Python、MATLAB、STM32、ANSYS、ROS2、SolidWorks，还是其他工程软件，都应采用统一的 Workflow 描述方式。

开发者描述的是工程，而不是代码。

Workflow 描述的是工程意图（Engineering Intent）。

后续所有 Compiler、Template、Generator、Runtime、Diagnose、Skill 均建立在 Workflow 之上。

因此，Workflow 不仅仅是一个可视化编辑器，更是 AIStudio 整个工程开发体系的起点，也是所有工程自动生成能力的基础。





第三章 Compiler 与 Execution Plan 系统设计

3.1 Compiler 设计背景

Compiler（编译器）是 AIStudio 的核心模块，也是整个系统的"大脑"。

在传统软件开发过程中，开发者编写源代码，再由 GCC、Clang、Javac 等编译器将代码转换成机器能够执行的程序。

而在 AIStudio 中，用户编写的并不是代码，而是 Workflow。因此，AIStudio 同样需要一个属于自己的 Compiler，用于将 Workflow 转换为平台能够理解的工程执行计划（Execution Plan）。

Compiler 的职责并不是生成代码，而是理解工作流、分析工作流、验证工作流，并将 Workflow 转换为标准化的工程描述，为后续 Generator 提供统一输入。

因此，Compiler 在 AIStudio 中的重要程度相当于传统编译器在程序开发中的作用。

3.2 Compiler 在系统中的定位

Compiler 位于 Workflow 与 Generator 之间，是连接可视化设计与工程生成的桥梁。

整个数据流必须严格遵循如下流程：

Workflow Editor │ ▼ Workflow Store │ ▼ workflow.json │ ▼ Compiler │ ▼ Execution Plan │ ▼ Generator │ ▼ Runtime 

Workflow 只负责描述工程。

Generator 只负责生成工程。

Runtime 只负责运行工程。

Compiler 是整个系统唯一允许解析 Workflow 的模块。

任何 Generator、Plugin、Skill 或 Runtime 都不得直接解析 workflow.json。

3.3 Compiler 的职责

Compiler 不负责生成 Python 代码。

Compiler 不负责运行工程。

Compiler 不负责调用 AI。

Compiler 仅负责完成以下工作：

读取 workflow.json。

解析整个工作流。

校验节点是否合法。

校验节点参数是否合法。

分析节点之间的依赖关系。

检查是否存在循环依赖。

检查插件是否存在。

检查模板是否存在。

检查 Generator 是否存在。

根据节点连接关系生成正确执行顺序。

生成统一 Execution Plan。

Compiler 永远不产生任何具体代码，它产生的是工程执行描述。

3.4 Compiler 工作流程

Compiler 的工作过程可以划分为多个阶段。

第一阶段，读取 Workflow。

Compiler 首先读取 workflow.json，并完成 JSON Schema 校验，确保工作流格式正确。

第二阶段，建立节点图。

根据 Nodes 与 Edges 建立完整有向图（Directed Graph）。

第三阶段，节点合法性检查。

检查节点是否存在。

节点类型是否合法。

节点参数是否完整。

插件是否加载成功。

第四阶段，依赖关系分析。

分析所有节点之间的数据流。

计算节点依赖。

建立节点关系树。

第五阶段，拓扑排序。

根据 Edge 自动完成 DAG（有向无环图）拓扑排序，得到正确执行顺序。

第六阶段，生成 Execution Plan。

将所有节点转换为统一执行计划。

最终交由 Generator 处理。

3.5 Workflow Graph

Compiler 在内部并不会直接操作 JSON。

读取 Workflow 后，应首先建立 Workflow Graph。

Workflow Graph 本质上是一张有向图。

图中的每一个 Node 表示一个工程节点。

每一条 Edge 表示节点之间的数据流。

例如：

Dataset │ ▼ YOLO │ ▼ LSTM │ ▼ Python Script 

Compiler 所有分析均基于 Workflow Graph 完成。

3.6 节点分析

Compiler 在遍历节点时，需要建立完整节点信息。

包括：

节点唯一 ID。

节点类型。

所属 Domain。

插件来源。

模板来源。

Generator 来源。

输入端口。

输出端口。

参数配置。

节点状态。

节点依赖。

运行优先级。

这些信息最终都会写入 Execution Plan。

3.7 参数校验

Compiler 必须负责参数合法性检查。

例如：

YOLO 节点：

Epoch 必须大于零。

Batch 必须合法。

Model 必须存在。

Dataset 必须存在。

MATLAB 节点：

模型名称不能为空。

仿真时间必须合法。

STM32 节点：

MCU 型号不能为空。

CubeMX 模板必须存在。

任何参数错误必须在 Compiler 阶段终止，而不是等 Generator 报错。

3.8 拓扑排序

Workflow 的执行顺序不能依赖节点摆放位置。

真正决定执行顺序的是 Edge。

Compiler 应采用 DAG 拓扑排序算法。

例如：

Dataset │ ▼ YOLO │ ▼ LSTM 

排序结果应为：

Dataset

↓

YOLO

↓

LSTM

如果发现：

A │ ▼ B ▲ │ 

形成闭环。

Compiler 必须立即报错：

Workflow 存在循环依赖。

禁止继续生成工程。

3.9 Domain 分发

AIStudio 支持多个专业领域。

Compiler 需要识别每一个节点所属 Domain。

例如：

AI。

MATLAB。

STM32。

ANSYS。

Python。

Compiler 根据 Domain 自动调用对应 Generator。

因此，Compiler 必须保持与 Domain 解耦。

以后增加新领域时，无需修改 Compiler 核心逻辑。

3.10 Execution Plan

Execution Plan 是 Compiler 唯一输出。

Execution Plan 并不是代码。

而是 Generator 可以直接理解的工程描述。

Execution Plan 至少包含：

执行顺序。

节点信息。

节点参数。

节点依赖。

模板路径。

Generator 类型。

Plugin 信息。

Domain 信息。

输入文件。

输出目录。

Execution Plan 应尽可能保持平台无关性，使不同 Generator 均可使用。

3.11 Compiler 与 Template 的关系

Compiler 不知道 Python 如何编写。

Compiler 不知道 MATLAB 如何编写。

Compiler 不知道 STM32 如何生成。

Compiler 只负责告诉 Generator：

当前需要生成什么工程。

需要哪些参数。

需要使用哪套模板。

真正工程生成由 Template Engine 完成。

因此：

Compiler 与 Template 必须完全解耦。

3.12 Compiler 与 Generator 的关系

Compiler 输出：

Execution Plan。

Generator 输入：

Execution Plan。

Generator 永远不能直接读取 workflow.json。

这样可以保证：

Workflow 数据结构修改时，只需要修改 Compiler，而 Generator 无需改动。

Execution Plan 将成为整个系统唯一标准接口。

3.13 Compiler 错误处理

Compiler 必须建立统一错误体系。

错误至少包括：

Workflow 格式错误。

节点不存在。

节点参数错误。

节点循环依赖。

模板不存在。

Generator 不存在。

插件未安装。

Domain 不支持。

所有错误均应统一返回 Diagnose Center。

不得直接弹窗。

不得直接终止程序。

3.14 Compiler 可扩展性

Compiler 必须采用插件化架构。

以后新增：

ROS2。

OpenCV。

TensorRT。

Unity。

Unreal。

无需修改 Compiler 核心代码。

新增 Domain 后，仅增加：

Domain Adapter。

Generator。

Template。

即可完成扩展。

Compiler 应始终保持稳定。

3.15 Compiler 与 AI

Compiler 不依赖 AI。

关闭 LLM 后：

Compiler 应完全正常工作。

AI 可以辅助：

Workflow 检查。

Workflow 优化。

节点推荐。

参数建议。

但 Compiler 的所有判断必须基于规则，而不是 AI 推理。

这样可以保证工程生成的稳定性和可重复性。

3.16 Compiler 最终目标

Compiler 是 AIStudio 最重要的基础设施之一。

它负责将可视化 Workflow 转换为标准化 Execution Plan，实现工作流与工程生成之间的彻底解耦。

未来，无论平台支持多少种专业软件、多少种模板、多少种 Generator，都无需修改 Workflow，只需通过 Compiler 完成统一编译，再交由对应 Generator 生成真实工程。

因此，Compiler 不只是一个 JSON 解析器，而是整个 AIStudio 的工程编译中心，也是平台能够持续扩展、支持多领域开发的关键核心模块。





第四章 Template Engine 与 Generator 系统设计

4.1 设计背景

在传统的低代码平台或代码生成平台中，大多数系统都是通过字符串拼接（String Concatenation）的方式生成代码。例如：

code += "model.train(" code += f"epochs={epoch}" code += ")" 

这种方式虽然实现简单，但随着业务复杂度增加，会出现代码难以维护、模板重复、可扩展性差、不同语言之间无法复用等问题。

AIStudio 不采用字符串拼接方式生成代码，而采用 Template（模板）驱动 的工程生成方式。

平台预先维护各类标准工程模板，Compiler 仅负责组织数据，Generator 根据 Execution Plan 调用对应模板，并完成变量填充，最终生成真实可运行的工程。

因此，Template Engine 是 AIStudio 工程生成体系的核心基础设施。

4.2 Template Engine 在系统中的定位

Template Engine 位于 Compiler 与 Generator 之间。

整体数据流如下：

Workflow │ ▼ workflow.json │ ▼ Compiler │ ▼ Execution Plan │ ▼ Template Engine │ ▼ Generator │ ▼ Real Project 

Template Engine 不负责解析 Workflow。

不负责运行工程。

不负责 AI 推理。

它唯一职责就是管理工程模板，并提供统一模板渲染能力。

4.3 为什么需要 Template Engine

AIStudio 希望支持：

Python

MATLAB

STM32CubeMX

ANSYS

ROS2

OpenCV

TensorRT

Unity

未来甚至支持更多专业软件。

如果每增加一个平台，都在 Generator 中增加大量 if-else：

if python if matlab if stm32 if ansys if ros 

Generator 将迅速膨胀，最终无法维护。

因此：

Generator 永远不关心代码内容。

Generator 只负责：

读取模板。

填充变量。

复制工程。

所有业务逻辑全部放入模板。

4.4 Template 的设计思想

Template 并不是一个代码片段。

而是一套完整工程。

例如：

YOLO 模板：

templates/ python/ yolo/ ├ train.py.tpl ├ predict.py.tpl ├ export.py.tpl ├ dataset.yaml.tpl ├ requirements.txt ├ README.md ├ .gitignore └ config.json 

Generator 不需要理解 train.py。

Generator 只需要：

复制整个目录。

替换变量。

即可得到完整工程。

4.5 Template 分类

AIStudio 所有模板按照 Domain 分类。

例如：

templates/ python/ matlab/ stm32/ ansys/ ros/ opencv/ common/ 

每一个 Domain 下面可以继续细分。

例如：

python/ yolo/ lstm/ classification/ segmentation/ detection/ custom-script/ 

这样方便以后插件动态增加模板。

4.6 Template 文件组成

一个标准 Template 至少包含：

工程目录。

源代码模板。

配置模板。

环境配置。

依赖配置。

启动脚本。

README。

License。

Git Ignore。

必要资源文件。

Generator 复制后即可直接运行。

4.7 Template 占位符

所有模板支持变量占位。

例如：

{{model}} {{epoch}} {{batch}} {{dataset}} {{output}} {{device}} {{learning_rate}} 

Generator 根据 Execution Plan 自动完成替换。

模板内部禁止硬编码业务参数。

4.8 Template Engine 工作流程

Template Engine 工作流程如下：

读取 Execution Plan。

定位 Template。

复制 Template。

扫描所有模板文件。

识别变量。

替换变量。

生成真实工程。

返回 Generator。

整个过程无需理解业务逻辑。

4.9 Generator 的定位

Generator 是整个工程生成模块。

Generator 不负责解析 Workflow。

Generator 不负责 Compiler。

Generator 不负责 Runtime。

Generator 唯一职责：

根据 Execution Plan 调用对应 Template。

生成真实工程。

因此：

Generator 可以理解为：

Template Dispatcher（模板调度器）。

4.10 Generator 分类

Generator 按照 Domain 分类。

例如：

Python Generator。

MATLAB Generator。

STM32 Generator。

ANSYS Generator。

ROS Generator。

每一个 Generator 均实现统一接口。

以后增加新的 Domain：

无需修改已有 Generator。

4.11 Generator 工作流程

Generator 执行流程如下：

读取 Execution Plan。

识别当前节点 Domain。

定位对应 Template。

调用 Template Engine。

生成工程目录。

输出工程。

整个过程不得直接拼接代码。

4.12 工程输出目录

所有 Generator 输出工程建议统一放置。

例如：

generated/ project-name/ python/ matlab/ stm32/ ansys/ 

这样一个 Workflow 可以同时生成多个专业工程。

例如：

同一个 Workflow：

既生成 Python AI 工程。

又生成 MATLAB 仿真工程。

又生成 STM32 控制工程。

实现真正跨平台开发。

4.13 多 Domain 工程生成

AIStudio 最大特点之一：

支持一个 Workflow 同时生成多个工程。

例如：

Dataset ↓ YOLO ↓ LSTM ↓ MATLAB ↓ STM32 

最终：

生成：

Python 工程。

MATLAB 工程。

STM32 工程。

Compiler 根据节点关系完成数据组织。

Generator 分别调用不同模板。

整个过程统一完成。

4.14 Template 与 Plugin

Template 支持插件扩展。

第三方开发者可以新增：

YOLO12。

SAM。

GroundingDINO。

ROS2。

OpenCV。

只需增加：

Template。

Generator。

Plugin。

无需修改平台源码。

因此：

Template 本身也是插件。

4.15 Template Version

所有 Template 必须拥有版本管理。

例如：

Version。

Author。

Create Time。

Support Domain。

Support Generator。

Compatible Compiler。

这样：

不同版本 Generator 可以自动选择兼容模板。

4.16 Template 校验

Generator 在生成工程之前必须完成：

模板存在检查。

模板版本检查。

变量完整检查。

模板文件完整检查。

依赖完整检查。

如果模板损坏：

禁止继续生成工程。

并返回 Diagnose Center。

4.17 Template 与 Runtime

Template 只负责生成工程。

Generator 只负责创建工程。

真正运行工程：

Runtime 完成。

Template 永远不调用 Python。

Template 永远不启动 MATLAB。

Template 与 Runtime 必须彻底解耦。

4.18 Template 与 AI

AI 不参与 Template。

AI 不负责生成代码。

AI 不修改模板。

AI 仅可：

推荐模板。

解释模板。

优化模板。

真正工程生成：

始终基于固定模板。

保证结果可重复。

4.19 Template Engine 最终目标

Template Engine 的最终目标，是建立一个面向工程开发的标准模板生态。

未来所有专业软件均通过统一模板描述工程，而不是通过字符串拼接生成代码。

Generator 永远保持简单。

Compiler 永远保持稳定。

Template 持续扩展。

开发者只需新增模板，即可让 AIStudio 支持新的开发平台。

因此，Template Engine 不只是一个模板替换工具，更是 AIStudio 能够支持 Python、MATLAB、STM32、ANSYS、ROS2 等多领域工程开发的基础设施，也是整个平台实现"一次设计，多工程生成"这一核心理念的关键。





第五章 Runtime、日志中心（Log Center）与 Diagnose Center 系统设计

5.1 设计背景

AIStudio 的目标不仅仅是帮助用户生成工程，更重要的是让用户能够在 AIStudio 内完成整个工程开发闭环。

传统低代码平台通常只能生成代码，随后需要用户自行打开 IDE、运行程序、查看终端、定位错误、修改代码，整个开发流程被割裂。

AIStudio 希望实现"工程生成→工程运行→日志采集→错误诊断→重新运行"的一体化开发体验，因此设计了 Runtime、Log Center 以及 Diagnose Center 三个核心模块。

其中：

Runtime 负责运行工程。

Log Center 负责采集和管理所有运行日志。

Diagnose Center 负责分析日志、定位问题，并在启用 AI 时调用大语言模型进行智能诊断。

这三个模块共同组成 AIStudio 的运行中心（Execution Center）。

5.2 Runtime 在系统中的定位

Runtime 位于 Generator 之后。

整个运行流程如下：

Workflow ↓ Compiler ↓ Execution Plan ↓ Generator ↓ Real Project ↓ Runtime ↓ Log Center ↓ Diagnose Center ↓ Skill（Optional） 

Runtime 是唯一允许启动本地程序的模块。

任何 Generator 不允许直接运行 Python。

任何 Workflow 不允许直接运行 MATLAB。

所有运行行为必须统一交给 Runtime。

5.3 Runtime 的职责

Runtime 的职责包括：

自动检测运行环境。

启动对应程序。

管理运行进程。

实时采集终端输出。

采集标准输出（stdout）。

采集错误输出（stderr）。

监控运行状态。

支持停止运行。

支持重新运行。

支持多任务运行。

支持运行完成通知。

支持运行失败通知。

Runtime 不负责解析错误。

Runtime 不负责 AI 推理。

Runtime 只负责"运行"。

5.4 多运行环境支持

AIStudio 不局限于 Python。

因此 Runtime 必须支持多个专业软件。

例如：

Python Runtime。

MATLAB Runtime。

STM32 Runtime。

ANSYS Runtime。

ROS Runtime。

未来可继续扩展：

SolidWorks。

OpenFOAM。

Unity。

Unreal。

每一种 Runtime 均实现统一接口。

以后增加新的运行环境，无需修改 Runtime 核心代码。

5.5 环境检测（Environment Detection）

在运行工程之前，Runtime 必须首先检测本地环境。

例如：

Python 是否安装。

Python 版本是否符合要求。

是否存在虚拟环境。

MATLAB 是否安装。

CubeMX 是否安装。

ANSYS 是否安装。

必要依赖是否完整。

如果环境缺失：

Runtime 不允许直接报错退出。

而是：

生成环境检测报告。

发送 Diagnose Center。

由 Diagnose Center 提示用户安装或修复。

5.6 工程运行

Runtime 根据工程类型自动选择运行方式。

例如：

Python 工程：

调用 Python Interpreter。

MATLAB 工程：

调用 MATLAB Engine 或命令行。

STM32 工程：

调用 CubeIDE 或编译工具链。

ANSYS 工程：

调用对应求解器。

整个运行过程应统一封装。

上层无需关心具体运行方式。

5.7 Process Manager（进程管理）

Runtime 内部建立统一进程管理器。

负责：

启动进程。

暂停进程。

恢复进程。

终止进程。

查询状态。

获取 PID。

监控 CPU。

监控内存。

监控运行时间。

以后支持多个工程同时运行。

每个运行任务均拥有独立 Process。

5.8 Terminal（终端）

AIStudio 内置统一终端。

所有程序运行输出均进入 Terminal。

包括：

Python 输出。

MATLAB 输出。

系统输出。

Generator 输出。

Compiler 输出。

Terminal 应支持：

彩色高亮。

自动滚动。

复制。

搜索。

保存日志。

清空日志。

过滤输出。

用户无需打开外部 CMD 或 Terminal。

5.9 Log Center

Log Center 是整个系统唯一日志中心。

所有模块均统一输出日志。

包括：

Compiler。

Generator。

Runtime。

Plugin。

Skill。

Provider。

Environment。

统一日志格式便于分析与检索。

禁止各模块自行打印日志。

5.10 日志分类

日志至少分为：

系统日志（System）。

运行日志（Runtime）。

编译日志（Compiler）。

Generator 日志。

Plugin 日志。

Environment 日志。

AI 日志。

错误日志（Error）。

警告日志（Warning）。

调试日志（Debug）。

信息日志（Info）。

日志应支持不同颜色显示。

方便开发者快速定位问题。

5.11 日志持久化

所有日志默认保存至项目目录。

例如：

logs/ compiler.log runtime.log generator.log system.log diagnose.log 

支持：

自动归档。

日志轮换。

日志压缩。

方便后续分析。

5.12 Diagnose Center

Diagnose Center 是整个平台的诊断中心。

它不负责运行程序。

它只负责分析问题。

Diagnose Center 接收：

Compiler 错误。

Generator 错误。

Runtime 错误。

Environment 错误。

Plugin 错误。

统一分析。

输出：

错误原因。

错误位置。

修复建议。

解决方案。

5.13 Diagnose 工作流程

Diagnose 工作流程如下：

Runtime 输出错误。

↓

Log Center 收集。

↓

Diagnose 接收日志。

↓

识别错误类型。

↓

生成诊断结果。

↓

展示给用户。

如果：

启用 AI。

↓

Skill 调用 LLM。

↓

返回更详细解释。

整个流程形成闭环。

5.14 Debug Skill

Diagnose Center 内置 Debug Skill。

Debug Skill 是 AIStudio 中最重要的 Skill 之一。

它负责：

读取错误日志。

分析报错。

定位错误。

生成修改建议。

支持：

Python Traceback。

ModuleNotFound。

ImportError。

CUDA Error。

MATLAB Error。

CubeMX Error。

ANSYS Error。

Debug Skill 不直接修改代码。

用户可以：

查看建议。

一键交给 AI 修改。

或者：

自行修改。

5.15 AI Optional

Diagnose 必须支持两种模式。

第一种：

不开启 AI。

Diagnose 根据规则：

分析日志。

提示错误。

第二种：

开启 AI。

Diagnose 将日志发送至：

LLM。

由 AI：

解释错误。

推荐修改。

优化 Workflow。

因此：

AI 永远只是增强能力。

Diagnose 必须可以独立工作。

5.16 日志可视化

Log Center 应支持：

实时刷新。

错误高亮。

搜索。

过滤。

日志等级。

时间排序。

模块排序。

点击日志。

自动定位对应节点。

未来支持：

日志时间轴。

运行历史。

性能分析。

5.17 Runtime 与 Workflow

Runtime 不读取前端。

Runtime 不解析 Workflow。

Runtime 只读取：

Generator 生成后的真实工程。

Generator 与 Runtime 必须彻底解耦。

Workflow 修改。

不会影响 Runtime。

5.18 Runtime 与 Plugin

插件允许扩展：

新的 Runtime。

例如：

Unity Runtime。

Docker Runtime。

ROS Runtime。

插件只需实现统一 Runtime Interface。

即可接入 AIStudio。

Runtime 保持稳定。

5.19 Runtime 最终目标

Runtime、Log Center 与 Diagnose Center 共同组成 AIStudio 的工程运行平台。

Generator 负责生成工程。

Runtime 负责运行工程。

Log Center 负责记录工程。

Diagnose Center 负责分析工程。

Skill 负责智能增强。

整个模块共同形成"生成—运行—记录—分析—优化"的完整工程闭环。

未来，无论支持多少种开发平台，都应采用统一 Runtime 接口、统一日志体系以及统一诊断体系，使 AIStudio 成为真正意义上的一站式工程开发平台，而不仅仅是一个代码生成工具。





第六章 Plugin、Skill、Provider 与平台扩展架构设计

6.1 设计背景

AIStudio 的目标不仅仅是完成一个 AI 工作流开发平台，更希望构建一个开放式、可扩展的工程开发生态。

传统开发软件通常采用固定功能设计，例如某一个软件只能完成 AI 训练，另一个软件只能完成 MATLAB 仿真，第三个软件只能完成嵌入式开发。当开发者需要支持新的框架、新的软件或者新的模型时，往往需要直接修改软件源码，维护成本高、扩展能力差。

因此，AIStudio 从设计之初便采用插件化（Plugin-Based Architecture）的思想，将节点、模板、Generator、AI 能力以及模型接入全部进行解耦，使平台能够随着插件不断扩展，而无需修改系统核心代码。

平台核心只负责提供基础运行框架，各种专业能力均以插件形式动态加载。

6.2 平台扩展理念

AIStudio 所有功能均遵循"核心稳定，能力扩展"的设计原则。

平台核心仅负责：

项目管理。

Workflow 编辑。

Compiler。

Template Engine。

Generator。

Runtime。

Log Center。

Diagnose Center。

除此之外，任何新的业务能力均建议通过插件进行扩展。

例如：

新增一个 YOLO12 节点。

新增一个 ROS2 Generator。

新增一个 STM32 模板。

新增一个 Gemini Provider。

新增一个 Debug Skill。

均无需修改平台源码。

平台只负责加载插件。

插件负责实现能力。

6.3 Plugin System

Plugin System 是 AIStudio 的插件管理中心。

负责整个插件生命周期管理。

包括：

插件发现。

插件安装。

插件卸载。

插件升级。

插件启用。

插件禁用。

插件版本管理。

插件依赖检查。

插件热加载。

插件冲突检测。

所有插件均由 Plugin Manager 统一管理。

禁止模块自行加载插件。

6.4 Plugin 分类

AIStudio 插件按照功能划分为多个类别。

第一类为 Node Plugin。

负责扩展新的 Workflow 节点。

例如：

YOLO。

LSTM。

SAM。

GroundingDINO。

OpenCV。

MATLAB。

STM32。

ANSYS。

每一个节点均可作为独立插件安装。

第二类为 Template Plugin。

负责提供工程模板。

例如：

Python Template。

MATLAB Template。

STM32CubeMX Template。

ANSYS Template。

Generator 在生成工程时自动调用对应模板。

第三类为 Generator Plugin。

负责新增新的工程生成器。

例如：

Python Generator。

ROS Generator。

Unity Generator。

TensorRT Generator。

Generator Plugin 只负责工程生成，不参与 Workflow。

第四类为 Runtime Plugin。

负责支持新的运行环境。

例如：

MATLAB Runtime。

Docker Runtime。

ROS Runtime。

Unity Runtime。

Runtime Plugin 实现统一运行接口。

第五类为 Skill Plugin。

负责扩展 AI 能力。

例如：

Workflow Planner。

Debug Skill。

Explain Skill。

Optimize Skill。

Environment Skill。

Skill Plugin 与系统运行完全解耦。

关闭 AI 后平台仍可正常运行。

第六类为 Provider Plugin。

负责接入不同的大语言模型。

例如：

OpenAI。

Anthropic Claude。

Google Gemini。

DeepSeek。

Qwen。

OpenAI Compatible API。

Provider Plugin 统一管理模型调用方式。

平台无需针对每一个模型分别开发。

6.5 Node Plugin

Node Plugin 是 Workflow 的组成部分。

每一个 Node Plugin 至少应提供：

节点名称。

节点类型。

节点图标。

节点分类。

输入端口。

输出端口。

参数定义。

参数校验规则。

默认模板。

默认 Generator。

帮助文档。

Node Plugin 不负责生成代码。

Node Plugin 仅负责描述节点能力。

6.6 Template Plugin

Template Plugin 提供完整工程模板。

每一个 Template Plugin 至少包含：

模板目录。

模板版本。

模板变量。

README。

依赖配置。

Generator 调用 Template Plugin 自动完成工程生成。

未来任何新框架均建议首先开发 Template Plugin。

6.7 Generator Plugin

Generator Plugin 负责：

读取 Execution Plan。

调用 Template。

填充变量。

生成工程。

Generator Plugin 不读取 Workflow。

Generator Plugin 不操作前端。

Generator Plugin 仅处理 Execution Plan。

所有 Generator 必须实现统一 Generator Interface。

6.8 Skill Plugin

Skill Plugin 是 AIStudio 的智能增强模块。

Skill 并不是系统运行的必要组成部分。

关闭 Skill 后：

Workflow。

Compiler。

Generator。

Runtime。

均应正常运行。

Skill 主要承担：

Workflow 自动生成。

自然语言规划。

错误解释。

日志分析。

参数推荐。

工程优化。

代码说明。

开发建议。

所有 Skill 必须基于统一 Skill Interface。

方便后续扩展。

6.9 Debug Skill

Debug Skill 是平台最重要的 Skill。

其职责包括：

读取 Runtime 日志。

解析 Compiler 错误。

解析 Python Traceback。

解析 MATLAB Error。

解析环境错误。

生成错误解释。

提供修复建议。

用户可选择：

查看建议。

自动交给 AI 修改。

或者自行修改。

Debug Skill 不直接修改工程。

而是提供辅助能力。

6.10 Provider Plugin

Provider Plugin 用于统一接入各种大语言模型。

AIStudio 不绑定任何模型厂商。

平台采用 Provider 抽象层。

任何支持 OpenAI API 格式的模型均可接入。

支持：

API Key。

Base URL。

Model Name。

Timeout。

Proxy。

Temperature。

Max Token。

用户可自由切换模型。

无需修改系统代码。

6.11 API Key 管理

AIStudio 提供统一 Provider 配置中心。

用户可以配置：

OpenAI。

Claude。

Gemini。

DeepSeek。

Qwen。

Moonshot。

以及其他兼容 OpenAI API 的模型。

API Key 应统一加密保存。

禁止硬编码。

支持：

新增。

编辑。

删除。

测试连接。

默认 Provider。

运行过程中按需调用。

6.12 MCP 扩展能力

AIStudio 支持 MCP（Model Context Protocol）扩展。

MCP 并不参与工程生成。

而是作为平台与外部软件通信的桥梁。

例如：

调用 MATLAB。

调用 ANSYS。

调用数据库。

调用浏览器。

调用企业内部系统。

Workflow 可通过 MCP 节点与外部软件建立连接，实现跨软件自动化。

6.13 Plugin 生命周期

所有插件均应遵循统一生命周期。

包括：

发现插件。

加载插件。

初始化。

注册能力。

运行。

暂停。

恢复。

卸载。

释放资源。

Plugin Manager 负责统一管理生命周期。

避免资源泄漏。

6.14 插件市场（Plugin Marketplace）

AIStudio 未来规划建立插件市场。

开发者可发布：

Node Plugin。

Template Plugin。

Generator Plugin。

Runtime Plugin。

Skill Plugin。

Provider Plugin。

用户可直接下载、安装、升级插件。

平台通过插件不断扩展能力，而无需频繁更新主程序。

6.15 SDK

AIStudio 提供 Plugin SDK。

开发者无需修改平台源码。

即可开发属于自己的插件。

SDK 应提供：

Plugin Interface。

Generator Interface。

Template Interface。

Runtime Interface。

Skill Interface。

Provider Interface。

Schema 定义。

开发者按照 SDK 即可完成插件开发。

6.16 AI Optional

AIStudio 坚持 AI Optional 设计理念。

AI 永远不是平台运行的必要条件。

关闭 AI 后：

Workflow。

Compiler。

Template。

Generator。

Runtime。

Log Center。

Diagnose。

全部仍可正常运行。

开启 AI 后：

仅增加：

智能规划。

日志解释。

自动推荐。

工作流生成。

智能优化。

因此：

AI 是平台能力增强层。

不是平台核心。

6.17 平台最终目标

Plugin、Skill、Provider 共同组成 AIStudio 的开放生态。

Workflow 描述工程。

Compiler 编译工程。

Generator 生成工程。

Runtime 运行工程。

Plugin 扩展工程。

Skill 增强工程。

Provider 提供 AI 能力。

整个系统形成"核心稳定、插件扩展、AI 增强"的整体架构，使 AIStudio 能够不断支持新的工程软件、新的开发框架以及新的 AI 模型，而无需重构平台核心，为未来长期演进提供稳定、开放且可持续发展的基础架构。






第七章 Project Manager、文件系统与工程生命周期设计

7.1 设计背景

AIStudio 并不是一个简单的网页应用，也不仅仅是一个 AI 聊天软件，而是一个真正面向工程开发的平台。因此，平台必须具备完整的项目（Project）管理能力，而不仅仅是保存一个 JSON 文件。

传统 IDE（如 Visual Studio、CLion、PyCharm、MATLAB、STM32CubeIDE）都以 Project（项目）作为管理单位，一个项目包含源代码、配置文件、资源文件、日志、输出结果等完整工程内容。

AIStudio 同样采用 Project 作为最小管理单元。Workflow、Template、Generator、Runtime、Log Center 等所有模块均围绕 Project 展开工作，而不是围绕单个 Workflow 文件。

Project Manager 的职责不仅是保存工程，更负责整个项目生命周期管理，包括创建、打开、保存、恢复、迁移、备份以及项目资源管理。

7.2 Project 在系统中的定位

Project 是 AIStudio 中最高层级的数据组织方式。

一个 Project 至少包含：

Workflow。

工程配置。

模板缓存。

生成工程。

运行日志。

插件配置。

数据集。

输出结果。

缓存文件。

整个系统遵循：

Project │ ▼ Workflow │ ▼ Compiler │ ▼ Generator │ ▼ Runtime 

任何 Workflow 必须属于某一个 Project。

任何 Generator 输出必须属于某一个 Project。

任何日志也必须归属于 Project。

7.3 Project 生命周期

一个 Project 从创建到结束，完整生命周期包括：

创建项目（Create）。

初始化目录。

创建默认 workflow.json。

创建 project.json。

初始化缓存。

打开项目（Open）。

恢复工作流。

恢复画布。

恢复最近运行记录。

恢复插件状态。

编辑项目（Edit）。

实时修改 Workflow。

更新配置。

更新资源。

生成工程。

运行工程。

保存项目（Save）。

同步 workflow.json。

同步 project.json。

更新资源。

写入日志。

关闭项目（Close）。

释放资源。

关闭 Runtime。

保存状态。

再次打开时恢复。

整个过程由 Project Manager 自动完成。

7.4 Project 目录结构

AIStudio 所有项目采用统一目录规范。

建议目录如下：

Project/ │ ├ workflow.json // 工作流 ├ project.json // 项目信息 ├ settings.json // 项目配置 │ ├ datasets/ // 数据集 ├ generated/ // 自动生成工程 ├ outputs/ // 输出结果 ├ logs/ // 日志 ├ cache/ // 缓存 ├ templates/ // 模板缓存 ├ plugins/ // 项目插件 ├ assets/ // 图片资源 ├ scripts/ // 用户脚本 ├ documents/ // 文档 └ temp/ // 临时文件 

所有模块均不得随意创建目录。

统一由 Project Manager 管理。

7.5 workflow.json

workflow.json 是整个项目最重要的数据文件。

它描述：

节点。

连线。

参数。

画布。

变量。

Domain。

Workflow Metadata。

Compiler 唯一读取：

workflow.json。

Generator 不读取前端。

Runtime 不读取 Workflow。

workflow.json 始终保持平台无关。

7.6 project.json

project.json 用于描述整个项目。

例如：

项目名称。

作者。

创建时间。

更新时间。

版本。

Compiler 版本。

Generator 版本。

Plugin 列表。

默认 Runtime。

默认 Provider。

最近打开时间。

支持 Domain。

Project Manager 根据 project.json 恢复整个项目。

7.7 项目配置

AIStudio 支持项目级配置。

包括：

默认 Python。

默认 MATLAB。

默认 Generator。

默认 Template。

默认 Runtime。

默认 Provider。

默认日志等级。

自动保存时间。

缓存目录。

生成目录。

项目配置与软件配置相互独立。

7.8 打开真实目录

AIStudio 必须支持打开电脑真实文件夹。

项目目录不是虚拟目录。

用户可以：

浏览本地目录。

创建新目录。

打开已有项目。

拖入项目。

拖入数据集。

直接访问真实文件系统。

未来：

支持：

Windows。

Linux。

macOS。

统一实现。

7.9 文件监听（File Watcher）

Project Manager 应建立文件监听系统。

监听：

workflow.json。

project.json。

datasets。

generated。

plugins。

如果外部修改：

Workflow。

自动刷新。

无需重新打开项目。

保持项目同步。

7.10 自动保存

AIStudio 支持自动保存。

Workflow 修改。

参数修改。

节点移动。

连线修改。

项目配置修改。

均自动保存。

采用防抖机制。

避免频繁写磁盘。

保证：

软件异常退出。

数据仍然完整。

7.11 最近项目

Project Manager 自动维护最近项目。

包括：

最近打开。

最近生成。

最近运行。

固定项目。

收藏项目。

支持：

一键重新打开。

方便开发者快速恢复工作。

7.12 多项目管理

AIStudio 后续支持：

同时打开多个 Project。

不同 Project：

拥有独立：

Workflow。

Generator。

Runtime。

Log。

Plugin。

避免互相影响。

Project Manager 统一调度。

7.13 工程生成目录

Generator 输出统一管理。

例如：

generated/ AI/ MATLAB/ STM32/ ANSYS/ 

不同 Domain：

互不影响。

方便：

重新生成。

重新运行。

删除旧工程。

7.14 数据集管理

Project Manager 提供数据集管理。

支持：

导入数据集。

删除数据集。

复制数据集。

检查数据集。

验证数据集。

以后：

支持：

YOLO。

COCO。

VOC。

ImageNet。

CSV。

Excel。

数据库。

统一管理。

7.15 输出管理

所有生成结果统一进入 outputs。

例如：

训练结果。

预测结果。

MATLAB 图片。

ANSYS 图片。

Excel。

CSV。

PDF。

模型权重。

统一管理。

方便：

版本控制。

历史查看。

导出。

分享。

7.16 缓存管理

Project Manager 管理：

Compiler Cache。

Generator Cache。

Template Cache。

Plugin Cache。

Log Cache。

用户可：

清理缓存。

重建缓存。

避免长期使用导致缓存膨胀。

7.17 备份与恢复

AIStudio 支持：

项目备份。

自动备份。

恢复备份。

历史版本。

以后：

支持：

Git。

云端同步。

版本比较。

Workflow Diff。

保证：

误删除。

仍可恢复。

7.18 Project 与 AI

Project Manager 不依赖 AI。

关闭 AI。

Project：

创建。

打开。

保存。

生成。

运行。

全部正常。

AI 仅提供：

项目说明。

目录分析。

Workflow 优化。

日志解释。

7.19 Project 最终目标

Project Manager 是 AIStudio 的基础设施之一，也是所有模块的统一入口。

Workflow、Compiler、Generator、Runtime、Log Center、Plugin、Skill 等模块均建立在 Project 基础之上。

通过统一的项目目录、统一的文件管理以及统一的生命周期管理，AIStudio 将传统分散的工程文件组织为标准化项目，使开发者能够以项目为中心管理整个工程开发流程。

Project 不仅是文件夹，更是整个 AIStudio 工程开发体系的组织核心，为未来团队协作、云端同步、版本管理、多工程开发以及插件生态提供统一基础。





第八章 Node Library、节点生态与 Workflow 编排系统设计

8.1 设计背景

AIStudio 的核心思想是以 Workflow 描述工程，而 Workflow 的基本组成单位就是 Node（节点）。

Node 不代表一段代码。

Node 代表一种工程能力（Engineering Capability）。

例如：

YOLO 节点代表目标检测训练。

LSTM 节点代表时间序列预测。

MATLAB 节点代表 MATLAB 仿真。

STM32 节点代表嵌入式工程。

ANSYS 节点代表有限元分析。

因此，Node Library（节点库）实际上就是 AIStudio 的能力中心。

未来平台支持什么软件、什么算法、什么业务，并不是修改系统源码，而是不断扩充节点库。

8.2 Node Library 在系统中的定位

Node Library 是整个 Workflow 的能力来源。

Workflow Editor 中所有可拖拽节点均来自 Node Library。

整个数据流如下：

Node Library │ ▼ Workflow Editor │ ▼ Workflow Store │ ▼ workflow.json │ ▼ Compiler │ ▼ Execution Plan 

因此：

Workflow 不负责定义节点。

Workflow 只负责使用节点。

Node Library 才是真正定义节点能力的地方。

8.3 Node 的设计理念

AIStudio 中每一个 Node 都应该满足以下原则：

第一，描述能力，而不是描述代码。

例如：

YOLO 节点描述的是：

"目标检测训练能力"

而不是：

train.py。

MATLAB 节点描述的是：

"MATLAB 仿真能力"

而不是：

main.m。

第二，节点必须平台无关。

例如：

Dataset 节点：

既可以提供给：

YOLO。

LSTM。

MATLAB。

也可以提供给：

STM32。

因此：

Node 不绑定具体 Generator。

第三，节点必须可组合。

任何节点均可自由组合。

例如：

Dataset ↓ YOLO ↓ LSTM ↓ Python Script ↓ MATLAB ↓ STM32 

平台不限制组合方式。

真正是否合法。

Compiler 判断。

8.4 Node 分类

AIStudio 官方节点库建议划分多个分类。

AI 节点

包括：

Dataset

YOLO

YOLO Export

Classification

Segmentation

Detection

Pose

OCR

LSTM

Transformer

Diffusion

LLM

Embedding

VectorDB

数据处理节点

包括：

CSV

Excel

JSON

XML

Image Loader

Video Loader

Database

HTTP

MQTT

Redis

Kafka

Python 节点

包括：

Python Script

Python Function

Python Package

Python Environment

Python Shell

MATLAB 节点

包括：

MATLAB Script

Simulink

Optimization

Signal Processing

Control

Image Processing

Machine Learning Toolbox

STM32 节点

包括：

CubeMX

GPIO

UART

PWM

ADC

CAN

FreeRTOS

Sensor

Motor

OTA

ANSYS 节点

包括：

Workbench

Mechanical

Fluent

APDL

Journal

Material

Mesh

Solver

控制流节点

包括：

If

Else

Switch

Loop

While

ForEach

Delay

Timer

Parallel

Merge

工具节点

包括：

Log

Debug

Print

Save

Export

Notification

Compress

Encrypt

MCP 节点

包括：

MCP Client

MCP Server

Browser

Database

Cloud

Local Tool

External Software

AI Skill 节点（可选）

包括：

Planner

Diagnose

Explain

Optimize

Environment

Generate Workflow

Auto Connect

8.5 Node 数据结构

每一个 Node 都应采用统一 Schema。

至少包含：

Node ID。

Node Type。

Node Category。

Node Version。

Node Icon。

Node Description。

Node Inputs。

Node Outputs。

Property Schema。

Validator。

Template Mapping。

Generator Mapping。

Plugin Source。

Support Domain。

Help Document。

这样：

Compiler 无需针对不同节点分别开发。

8.6 Port（端口）设计

每一个 Node 都拥有输入端口与输出端口。

端口不仅用于连接节点。

更代表数据流。

例如：

Dataset：

输出：

Image Dataset。

YOLO：

输入：

Dataset。

输出：

Weight。

Prediction。

这样：

Compiler 能够根据 Port 自动判断连接是否合法。

未来支持：

数据类型检查。

端口兼容检查。

自动类型转换。

8.7 Node Property

每一个节点均拥有独立 Property。

例如：

YOLO：

Model。

Epoch。

Batch。

Learning Rate。

Image Size。

Workers。

Device。

MATLAB：

Simulation Time。

Solver。

Step。

Model。

STM32：

MCU。

Clock。

RTOS。

Compiler。

所有 Property：

统一由 Property Schema 描述。

Workflow Editor 自动生成属性面板。

无需手写 UI。

8.8 Node Validator

每一个 Node 均拥有 Validator。

例如：

YOLO：

Epoch > 0。

Batch > 0。

Model 不为空。

Dataset 已连接。

MATLAB：

模型存在。

Solver 合法。

Compiler 调用 Validator。

而不是：

Generator。

这样：

错误能够提前发现。

8.9 Node 与 Template

Node 不包含代码。

Node 只描述能力。

真正工程生成：

依赖：

Template。

例如：

YOLO：

对应：

YOLO Template。

MATLAB：

对应：

MATLAB Template。

Generator 自动查找。

Node 不直接生成任何代码。

8.10 Node 与 Generator

Node 不知道 Generator。

Generator 根据：

Template Mapping。

自动寻找 Generator。

例如：

YOLO：

Generator：

Python Generator。

MATLAB：

Generator：

MATLAB Generator。

这样：

Node 与 Generator 解耦。

8.11 Node 与 Plugin

未来所有 Node：

均支持插件扩展。

开发者只需：

新增：

Node Plugin。

即可增加：

新节点。

无需修改平台源码。

例如：

SAM。

YOLO12。

GroundingDINO。

OpenCV。

TensorRT。

均可作为：

Node Plugin。

8.12 Node Search

节点库应支持：

搜索。

分类。

收藏。

最近使用。

标签。

拼音搜索。

英文搜索。

模糊搜索。

方便：

大型节点库快速查找。

8.13 Node Marketplace

未来建立节点市场。

第三方开发者：

上传：

Node。

Template。

Generator。

Plugin。

用户：

下载。

安装。

评分。

评论。

升级。

AIStudio 官方节点库持续扩展。

8.14 Workflow 编排

Workflow 本质上就是：

Node 编排。

而不是：

代码编排。

Workflow：

通过：

Node。

Port。

Edge。

完成整个工程描述。

Compiler：

解析 Workflow。

Generator：

生成工程。

Node：

永远只描述能力。

8.15 Node 与 AI

AI 可以：

推荐节点。

自动生成 Workflow。

自动连接节点。

解释节点。

生成参数。

但是：

Node 本身：

不依赖 AI。

关闭 AI：

节点仍可正常工作。

8.16 Node Library 最终目标

Node Library 是 AIStudio 最重要的能力中心。

未来 AIStudio 所支持的所有算法、所有工程软件、所有开发框架，都将以 Node 的形式接入平台。

开发者不再学习不同软件的开发流程，而是学习如何组合 Node。

Workflow 不断扩展。

Node Library 持续增长。

Plugin 持续丰富。

最终形成覆盖 AI、控制、仿真、嵌入式、工业软件等多个领域的统一节点生态，使 AIStudio 真正成为一个面向工程开发的 Workflow 平台，而不是单一领域的低代码工具。






第九章 LLM、AI Assistant 与智能工作流系统设计

9.1 设计背景

AIStudio 并不仅仅是一个可视化 Workflow 平台，它同时也是一个基于大语言模型（Large Language Model，LLM）的智能工程开发助手。

但是，AIStudio 的定位与传统 AI 聊天工具不同。

ChatGPT、Claude、Gemini 等产品主要以自然语言问答为核心，而 AIStudio 更关注工程开发过程中的智能辅助能力，例如工作流规划、参数推荐、日志分析、工程调试和代码解释。

因此，在 AIStudio 中，LLM 并不是平台的核心，而是平台的智能增强层（Intelligence Layer）。

平台必须保证：即使完全关闭 LLM，Workflow、Compiler、Generator、Runtime 等核心模块仍然能够正常运行；开启 LLM 后，仅在原有基础上增加智能能力，而不会改变系统架构。

9.2 AI Assistant 的定位

AI Assistant 是用户与平台之间的智能交互入口。

用户既可以通过拖拽节点完成 Workflow，也可以直接使用自然语言描述需求。

AI Assistant 负责理解用户需求，并将自然语言转换为平台能够理解的工作流描述。

整个交互流程如下：

User │ ▼ AI Assistant │ ▼ Workflow Planner Skill │ ▼ Workflow Draft │ ▼ Workflow Editor 

AI Assistant 不直接生成 Python 代码，也不直接修改工程，而是负责理解需求并组织 Workflow。

9.3 AI Optional 设计理念

AIStudio 采用 AI Optional 设计思想。

LLM 永远不是系统运行的必要条件。

关闭 AI 后：

Workflow 可以编辑。

Compiler 可以编译。

Generator 可以生成工程。

Runtime 可以运行工程。

Diagnose 可以进行规则诊断。

开启 AI 后：

自动规划 Workflow。

自动连接节点。

推荐参数。

解释错误。

优化工程。

自动生成文档。

因此，AI 是平台能力增强层，而不是平台运行基础。

9.4 LLM Provider 抽象层

为了避免平台绑定某一家模型厂商，AIStudio 在架构中引入 Provider 抽象层。

所有模型均通过统一接口接入，例如：

OpenAI

Claude

Gemini

DeepSeek

Qwen

Kimi

OpenAI Compatible API

Provider 负责：

API 调用。

Token 管理。

上下文管理。

流式输出。

超时控制。

错误处理。

AI Assistant 永远只调用 Provider Interface，而不直接调用具体厂商 API。

9.5 API Key 管理

平台提供统一的模型配置中心。

用户可以配置：

Provider 名称。

Base URL。

API Key。

默认模型。

Temperature。

Max Tokens。

Top P。

System Prompt。

API Key 应加密保存在本地配置文件中，不允许写入 Workflow 或项目文件。

不同项目可共享 Provider，也可配置独立 Provider。

9.6 Workflow Planner Skill

Workflow Planner 是 AIStudio 最核心的 Skill。

它负责将用户的自然语言需求转换为 Workflow。

例如：

用户输入：

"帮我训练一个 YOLO 模型，然后用 LSTM 分析交通流，最后生成 MATLAB 仿真。"

Workflow Planner 输出：

Dataset │ ▼ YOLO │ ▼ LSTM │ ▼ MATLAB 

同时自动填写节点参数，并连接端口。

Planner 不生成代码，而是生成 Workflow。

9.7 Auto Connect（自动连线）

AI Assistant 可以根据节点输入输出端口自动建立连接。

例如：

Dataset 输出 Image。

YOLO 输入 Image。

Planner 自动连接。

如果端口类型不兼容：

Planner 自动提示：

无法建立连接。

避免用户手动寻找端口。

9.8 Parameter Recommendation（参数推荐）

AI Assistant 可以根据任务推荐节点参数。

例如：

YOLO：

推荐 Epoch。

推荐 Batch。

推荐 Learning Rate。

推荐模型版本。

LSTM：

推荐 Time Window。

推荐 Hidden Size。

推荐 Optimizer。

MATLAB：

推荐 Solver。

推荐 Simulation Time。

推荐参数仅作为建议。

最终由用户确认。

9.9 Workflow Explain

AI Assistant 可以解释整个 Workflow。

例如：

用户点击：

"解释工作流"

AI 自动说明：

每个节点作用。

数据流方向。

输入输出关系。

最终生成内容。

帮助初学者理解整个工程。

9.10 Debug Skill

当 Runtime 出现错误时：

Log Center：

采集日志。

Diagnose：

识别错误。

Debug Skill：

结合日志生成：

错误原因。

可能原因。

解决建议。

必要时调用 LLM：

进一步分析复杂错误。

例如：

Python Traceback。

CUDA Error。

ImportError。

MATLAB Error。

CubeMX Error。

均可分析。

9.11 Environment Skill

Environment Skill 用于环境诊断。

负责：

检测 Python。

检测 CUDA。

检测 MATLAB。

检测 CubeMX。

检测依赖。

检测插件。

如果发现缺失：

自动提示：

缺少 Python。

缺少 pip。

缺少 CUDA。

缺少插件。

必要时提供安装建议。

9.12 Explain Skill

Explain Skill 用于知识解释。

例如：

解释：

YOLO。

LSTM。

Transformer。

MCP。

Workflow。

Execution Plan。

Generator。

帮助用户快速学习平台。

9.13 Optimize Skill

Optimize Skill 用于优化 Workflow。

例如：

发现重复节点。

发现无效连接。

发现性能瓶颈。

推荐更优节点。

推荐更优模板。

推荐更优 Generator。

帮助开发者持续优化工程。

9.14 Prompt 管理

AIStudio 不应将 Prompt 写死在代码中。

所有 Prompt 建议统一管理。

例如：

Workflow Planner Prompt。

Debug Prompt。

Explain Prompt。

Optimize Prompt。

Environment Prompt。

Prompt 可独立维护、更新和版本管理。

9.15 Context 管理

为了保证 LLM 输出质量，AIStudio 应建立统一 Context 管理系统。

上下文包括：

当前 Workflow。

当前节点。

当前日志。

当前项目。

当前插件。

当前 Provider。

不同 Skill 根据需要选择上下文，而不是一次发送全部内容，以降低 Token 消耗。

9.16 Token 管理

平台应统计：

本次 Token。

今日 Token。

Provider Token。

缓存命中。

请求耗时。

方便开发者了解 AI 使用情况。

未来支持：

Token 配额。

费用统计。

Provider 对比。

9.17 LLM 与 Workflow 的关系

LLM 不直接修改 Workflow。

LLM 输出的是：

Workflow Draft。

最终：

Workflow Editor：

展示。

用户：

确认。

Workflow Store：

写入。

保证：

AI 不会直接修改项目。

用户始终拥有最终控制权。

9.18 AI Assistant 最终目标

AI Assistant 并不是一个简单的聊天机器人，而是 AIStudio 的智能工程助手。

它能够理解开发需求、规划工作流、推荐参数、分析日志、解释系统、优化工程，并与 Workflow、Compiler、Generator、Runtime 等模块形成完整协作关系。

最终，AIStudio 希望实现"自然语言 + 可视化工作流"双模式开发，让专业开发者能够获得更高效率，也让非专业开发者能够通过 AI 辅助快速完成复杂工程，实现真正意义上的智能工程开发平台。





第十章 平台未来规划、生态建设与发展路线图

10.1 设计目标

AIStudio 的目标并不仅仅是完成一个 AI 模型训练平台，也不仅仅是一个工作流软件，而是打造一个面向人工智能、科学计算、工业仿真、嵌入式开发和智能制造的一站式工程开发平台。

平台希望打破不同开发工具之间的数据壁垒，通过统一的 Workflow、统一的数据描述（JSON）、统一的工程生成机制（Template + Generator），让开发者可以使用一种开发方式完成多个专业领域的工程构建。

未来，无论是 AI 算法工程师、自动驾驶工程师、机器人开发者、嵌入式工程师，还是科研人员，都能够在 AIStudio 中完成自己的开发任务。

10.2 平台发展阶段

为了保证平台持续演进，AIStudio 的发展规划划分为四个阶段。

第一阶段：AI 工作流平台（当前阶段）

完成 Workflow 编辑器。

完成 JSON 工作流。

完成 Compiler。

完成 Template Engine。

完成 Generator。

完成 Runtime。

支持 Python AI 工程生成。

实现 YOLO、LSTM 等基础 AI 工作流。

第二阶段：多领域工程平台

增加：

MATLAB。

Simulink。

STM32CubeMX。

ANSYS。

ROS2。

OpenCV。

Docker。

实现真正跨平台工程生成。

Workflow 可以同时生成多个专业工程。

第三阶段：插件生态平台

上线：

Plugin Marketplace。

Skill Marketplace。

Template Marketplace。

Node Marketplace。

第三方开发者可发布：

节点。

模板。

Generator。

Skill。

Runtime。

形成开放生态。

第四阶段：云平台与团队协作

支持：

云端 Workflow。

团队协同开发。

版本管理。

多人实时编辑。

在线运行。

模型管理。

插件云同步。

实现 AIStudio Cloud。

10.3 官方节点生态

未来官方节点库持续扩充。

计划支持：

AI。

机器人。

工业控制。

自动驾驶。

数字孪生。

机械设计。

建筑。

交通。

电力。

材料。

生物医学。

未来节点数量预计超过数百个。

开发者无需编写底层代码，仅通过 Workflow 即可完成复杂工程搭建。

10.4 官方模板生态

平台将持续维护官方 Template。

例如：

YOLO Template。

LSTM Template。

MATLAB Template。

STM32 Template。

ROS2 Template。

TensorRT Template。

Unity Template。

用户始终能够使用官方最佳实践生成标准工程。

10.5 官方 Skill 生态

官方 Skill 持续扩展。

例如：

Workflow Planner。

Debug。

Explain。

Optimize。

Environment。

Document Generator。

Requirement Analyzer。

Architecture Designer。

未来每一个 Skill 都可单独升级。

10.6 官方 Provider 支持

未来平台将持续适配国内外主流模型。

包括：

OpenAI。

Claude。

Gemini。

DeepSeek。

Qwen。

Kimi。

Llama。

以及任何兼容 OpenAI API 的模型。

用户拥有完全自由的模型选择权。

10.7 MCP 生态

未来 AIStudio 将深度支持 MCP（Model Context Protocol）。

实现：

MATLAB。

ANSYS。

浏览器。

数据库。

Office。

企业内部软件。

实验设备。

工业控制系统。

全部通过 MCP 接入。

Workflow 不仅能够控制 AI，也能够控制真实软件和设备。

10.8 云端能力

未来支持：

项目同步。

Workflow 云存储。

模型云训练。

插件云下载。

模板云同步。

日志云分析。

AI 在线推理。

用户无需依赖单台电脑即可完成开发。

10.9 团队协作

未来支持：

多人协同 Workflow 编辑。

权限管理。

成员管理。

评论。

任务分配。

版本比较。

冲突合并。

让 AIStudio 从个人工具发展为团队开发平台。

10.10 开放 SDK

官方提供完整 SDK。

包括：

Plugin SDK。

Node SDK。

Generator SDK。

Runtime SDK。

Skill SDK。

Provider SDK。

Template SDK。

第三方开发者可快速开发属于自己的扩展能力。

10.11 开源生态

未来计划逐步开放部分核心组件。

例如：

Workflow Schema。

Plugin SDK。

Template SDK。

Node SDK。

Compiler Interface。

吸引开发者共同完善平台。

形成健康的开源生态。

10.12 科研应用

AIStudio 不仅服务工业开发，也服务科研工作。

未来支持：

论文实验复现。

算法验证。

科研 Workflow 保存。

实验记录。

实验结果管理。

帮助研究人员快速完成实验设计与验证，提高科研效率。

10.13 教学应用

未来可应用于高校教学。

例如：

人工智能课程。

Python 编程课程。

MATLAB 建模课程。

STM32 嵌入式课程。

自动控制课程。

学生无需复杂环境配置，即可通过 Workflow 学习工程开发流程。

10.14 工业应用

未来可扩展至工业领域。

例如：

自动驾驶算法开发。

智能制造。

机器人控制。

数字孪生。

智能交通。

工业视觉检测。

预测性维护。

帮助企业降低开发成本，提高跨团队协作效率。

10.15 平台愿景

AIStudio 希望成为连接人工智能、工程软件与工业开发的统一平台。

通过 Workflow、JSON、Compiler、Template、Generator、Runtime、Plugin、Skill、LLM 等模块，将不同领域的软件开发方式统一到一个开放、标准、可扩展的架构之中。

未来，开发者不再需要分别学习多个专业软件的工程组织方式，而是通过统一的工作流描述工程，通过统一的平台生成工程，通过统一的运行环境验证工程，通过统一的 AI 助手优化工程。

AIStudio 最终希望实现的，不只是"低代码开发"，也不只是"AI 编程"，而是建立一种全新的工程开发范式——以工作流为核心、以 JSON 为桥梁、以插件为生态、以 AI 为增强的一站式智能工程开发平台。